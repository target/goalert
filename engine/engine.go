package engine

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/target/goalert/alert"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/engine/cleanupmanager"
	"github.com/target/goalert/engine/escalationmanager"
	"github.com/target/goalert/engine/heartbeatmanager"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/engine/npcyclemanager"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/engine/rotationmanager"
	"github.com/target/goalert/engine/schedulemanager"
	"github.com/target/goalert/engine/statusupdatemanager"
	"github.com/target/goalert/engine/verifymanager"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

type updater interface {
	Name() string
	UpdateAll(context.Context) error
}

// Engine handles automatic escaltion of unacknowledged(triggered) alerts, as well as
// passing to-be-sent notifications to the notification.Sender.
//
// Care is taken to ensure only one attempt is made per contact-method
// at a time, regardless of how many instances of the application may be running.
type Engine struct {
	b   *backend
	mgr *lifecycle.Manager

	shutdownCh  chan struct{}
	triggerCh   chan struct{}
	runLoopExit chan struct{}

	nextCycle chan chan struct{}

	modules []updater
	msg     *message.DB

	am  alert.Manager
	cfg *Config

	triggerPauseCh chan *pauseReq
}

type pauseReq struct {
	ch  chan error
	ctx context.Context
}

// NewEngine will create a new Engine using the passed *sql.DB as a backend. Outgoing
// notifications will be passed to Sender.
//
// Context is only used for preparing and initializing.
func NewEngine(ctx context.Context, db *sql.DB, c *Config) (*Engine, error) {
	var err error

	p := &Engine{
		cfg:            c,
		shutdownCh:     make(chan struct{}),
		triggerCh:      make(chan struct{}),
		triggerPauseCh: make(chan *pauseReq),
		runLoopExit:    make(chan struct{}),
		nextCycle:      make(chan chan struct{}),

		am: c.AlertStore,
	}

	p.mgr = lifecycle.NewManager(p._run, p._shutdown)
	err = p.mgr.SetPauseResumer(lifecycle.PauseResumerFunc(
		p._pause,
		p._resume,
	))
	if err != nil {
		return nil, err
	}

	rotMgr, err := rotationmanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "rotation management backend")
	}
	schedMgr, err := schedulemanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "schedule management backend")
	}
	epMgr, err := escalationmanager.NewDB(ctx, db, c.AlertLogStore)
	if err != nil {
		return nil, errors.Wrap(err, "alert escalation backend")
	}
	ncMgr, err := npcyclemanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "notification cycle backend")
	}
	statMgr, err := statusupdatemanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "status update backend")
	}
	verifyMgr, err := verifymanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "verification backend")
	}
	hbMgr, err := heartbeatmanager.NewDB(ctx, db, c.AlertStore)
	if err != nil {
		return nil, errors.Wrap(err, "heartbeat processing backend")
	}
	cleanMgr, err := cleanupmanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "cleanup backend")
	}

	p.modules = []updater{
		rotMgr,
		schedMgr,
		epMgr,
		ncMgr,
		statMgr,
		verifyMgr,
		hbMgr,
		cleanMgr,
	}

	p.msg, err = message.NewDB(ctx, db, &message.Config{
		MaxMessagesPerCycle: c.MaxMessages,
		RateLimit: map[notification.DestType]*message.RateConfig{
			notification.DestTypeSMS:          &message.RateConfig{PerSecond: 1, Batch: 5 * time.Second},
			notification.DestTypeVoice:        &message.RateConfig{PerSecond: 1, Batch: 5 * time.Second},
			notification.DestTypeSlackChannel: &message.RateConfig{PerSecond: 5, Batch: 5 * time.Second},
		},
		Pausable: p.mgr,
	})
	if err != nil {
		return nil, errors.Wrap(err, "messaging backend")
	}

	p.b, err = newBackend(db)
	if err != nil {
		return nil, errors.Wrap(err, "init backend")
	}

	return p, nil
}

// WaitNextCycle will return after the next engine cycle starts and then finishes.
func (p *Engine) WaitNextCycle(ctx context.Context) error {
	select {
	case ch := <-p.nextCycle:
		select {
		case <-ch:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Engine) processModule(ctx context.Context, m updater) {
	defer recoverPanic(ctx, m.Name())
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		err := m.UpdateAll(ctx)
		if sqlErr := sqlutil.MapError(err); ctx.Err() == nil && sqlErr != nil && strings.HasPrefix(sqlErr.Code, "40") {
			// Class `40` is a transaction failure.
			// In that case we will retry, so long
			// as the context deadline has not been reached.
			//
			// https://www.postgresql.org/docs/9.6/static/errcodes-appendix.html
			continue
		}
		if err != nil && errors.Cause(err) != processinglock.ErrNoLock {
			log.Log(ctx, errors.Wrap(err, m.Name()))
		}
		break
	}
}

func (p *Engine) processMessages(ctx context.Context) {
	ctx, sp := trace.StartSpan(ctx, "Engine.MessageManager")
	defer sp.End()
	defer recoverPanic(ctx, "MessageManager")
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err := p.msg.SendMessages(ctx, p.sendMessage, p.cfg.NotificationSender.Status)
	if errors.Cause(err) == processinglock.ErrNoLock {
		return
	}
	if errors.Cause(err) == message.ErrAbort {
		return
	}
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "send outgoing messages"))
	}
}

func recoverPanic(ctx context.Context, name string) {
	err := recover()
	if err == nil {
		return
	}

	if e, ok := err.(error); ok {
		log.Log(ctx, errors.Wrapf(e, "PANIC in %s", name))
	} else {
		log.Log(ctx, errors.Errorf("PANIC in %s: %+v", name, err))
	}
}

// Trigger will force notifications to be processed immediately.
func (p *Engine) Trigger() {
	<-p.triggerCh
}

// Pause will attempt to gracefully stop engine processing.
func (p *Engine) Pause(ctx context.Context) error {
	ctx, sp := trace.StartSpan(ctx, "Engine.Pause")
	defer sp.End()

	return p.mgr.Pause(ctx)
}
func (p *Engine) _pause(ctx context.Context) error {
	ch := make(chan error, 1)

	select {
	case <-p.shutdownCh:
		return errors.New("shutting down")
	case <-ctx.Done():
		return ctx.Err()
	case p.triggerPauseCh <- &pauseReq{ch: ch, ctx: ctx}:
		select {
		case <-ctx.Done():
			defer p.Resume(ctx)
			return ctx.Err()
		case err := <-ch:
			return err
		}
	}
}

// Resume will allow the engine to resume processing.
func (p *Engine) Resume(ctx context.Context) error {
	return p.mgr.Resume(ctx)
}
func (p *Engine) _resume(ctx context.Context) error {
	// nothing to be done `p.mgr.IsPaused` will already
	// return false
	return nil
}

// Run will being the engine loop.
func (p *Engine) Run(ctx context.Context) error {
	return p.mgr.Run(ctx)
}

// Shutdown will gracefully shutdown the processor, finishing any ongoing tasks.
func (p *Engine) Shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}
	ctx, sp := trace.StartSpan(ctx, "Engine.Shutdown")
	defer sp.End()

	return p.mgr.Shutdown(ctx)
}
func (p *Engine) _shutdown(ctx context.Context) error {
	close(p.shutdownCh)
	<-p.runLoopExit
	return nil
}

// UpdateStatus will update the status of a message.
func (p *Engine) UpdateStatus(ctx context.Context, status *notification.MessageStatus) error {
	var err error
	permission.SudoContext(ctx, func(ctx context.Context) {
		err = p.msg.UpdateMessageStatus(ctx, status)
	})
	return err
}

// Receive will process a notification result.
func (p *Engine) Receive(ctx context.Context, callbackID string, result notification.Result) error {
	ctx, sp := trace.StartSpan(ctx, "Engine.Receive")
	defer sp.End()
	cb, err := p.b.FindOne(ctx, callbackID)
	if err != nil {
		return err
	}

	var usr *user.User
	permission.SudoContext(ctx, func(ctx context.Context) {
		cm, serr := p.cfg.ContactMethodStore.FindOne(ctx, cb.ContactMethodID)
		if serr != nil {
			err = errors.Wrap(serr, "lookup contact method")
			return
		}
		usr, serr = p.cfg.UserStore.FindOne(ctx, cm.UserID)
		if serr != nil {
			err = errors.Wrap(serr, "lookup user")
		}
	})
	if err != nil {
		return err
	}
	ctx = permission.UserSourceContext(ctx, usr.ID, usr.Role, &permission.SourceInfo{
		Type: permission.SourceTypeNotificationCallback,
		ID:   callbackID,
	})

	switch result {
	case notification.ResultAcknowledge:
		return p.am.UpdateStatus(ctx, cb.AlertID, alert.StatusActive)
	case notification.ResultResolve:
		return p.am.UpdateStatus(ctx, cb.AlertID, alert.StatusClosed)
	}

	return errors.New("unknown result")
}

// Stop will disable all associated contact methods associated with `value` of type `t`. This is should
// be invoked if a user, for example, responds with `STOP` via SMS.
func (p *Engine) Stop(ctx context.Context, d notification.Dest) error {
	if !d.Type.IsUserCM() {
		return errors.New("stop only supported on user contact methods")
	}
	var err error

	permission.SudoContext(ctx, func(ctx context.Context) {
		err = p.cfg.ContactMethodStore.DisableByValue(ctx, contactmethod.TypeFromDestType(d.Type), d.Value)
	})

	return err
}

func (p *Engine) processAll(ctx context.Context) bool {
	for _, m := range p.modules {
		if p.mgr.IsPausing() {
			return true
		}
		ctx, sp := trace.StartSpan(ctx, m.Name())
		p.processModule(ctx, m)
		sp.End()
	}
	return false
}
func (p *Engine) cycle(ctx context.Context) {
	ctx, sp := trace.StartSpan(ctx, "Engine.Cycle")
	defer sp.End()

	ctx = p.cfg.ConfigSource.Config().Context(ctx)

	ch := make(chan struct{})
	defer close(ch)
passSignals:
	for {
		select {
		case p.nextCycle <- ch:
		default:
			break passSignals
		}
	}

	if p.mgr.IsPausing() {
		log.Logf(ctx, "Engine cycle disabled (paused or shutting down).")
		sp.AddAttributes(trace.BoolAttribute("cycle.skip", true))
		return
	}

	log.Logf(ctx, "Engine cycle start.")
	defer log.Logf(ctx, "Engine cycle end.")

	aborted := p.processAll(ctx)
	if aborted || p.mgr.IsPausing() {
		sp.Annotate([]trace.Attribute{trace.BoolAttribute("cycle.abort", true)}, "Cycle aborted.")
		log.Logf(ctx, "Engine cycle aborted (paused or shutting down).")
		return
	}
	p.processMessages(ctx)
}
func (p *Engine) handlePause(ctx context.Context, respCh chan error) {
	// nothing special to do currently
	respCh <- nil
}

func (p *Engine) _run(ctx context.Context) error {
	defer close(p.runLoopExit)
	ctx = permission.SystemContext(ctx, "Engine")
	if p.cfg.DisableCycle {
		log.Logf(ctx, "Engine started in API-only mode.")
		for {
			select {
			case req := <-p.triggerPauseCh:
				req.ch <- nil
			case <-ctx.Done():
				return ctx.Err()
			case <-p.shutdownCh:
				return nil
			case p.triggerCh <- struct{}{}:
				log.Logf(ctx, "Ignoring engine trigger (API-only mode).")
			}
		}
	}

	alertTicker := time.NewTicker(5 * time.Second)
	defer alertTicker.Stop()

	defer close(p.triggerCh)

	p.cycle(ctx)

	for {
		// give priority to pending shutdown signals
		// otherwise if the processing loop takes longer than
		// 5 seconds, it may never shut down.
		select {
		case req := <-p.triggerPauseCh:
			p.handlePause(req.ctx, req.ch)
		case <-ctx.Done():
			// run context canceled or something
			return ctx.Err()
		case <-p.shutdownCh:
			// shutdown requested
			return nil
		default:
		}

		select {
		case req := <-p.triggerPauseCh:
			p.handlePause(req.ctx, req.ch)
		case p.triggerCh <- struct{}{}:
			p.cycle(log.WithField(ctx, "Trigger", "DIRECT"))
		case <-alertTicker.C:
			p.cycle(log.WithField(ctx, "Trigger", "INTERVAL"))
		case <-ctx.Done():
			// context canceled or something
			return ctx.Err()
		case <-p.shutdownCh:
			// shutdown requested
			return nil
		}
	}
}
