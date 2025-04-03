package engine

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/auth/authlink"
	"github.com/target/goalert/engine/cleanupmanager"
	"github.com/target/goalert/engine/compatmanager"
	"github.com/target/goalert/engine/escalationmanager"
	"github.com/target/goalert/engine/heartbeatmanager"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/engine/metricsmanager"
	"github.com/target/goalert/engine/npcyclemanager"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/engine/rotationmanager"
	"github.com/target/goalert/engine/schedulemanager"
	"github.com/target/goalert/engine/signalmgr"
	"github.com/target/goalert/engine/statusmgr"
	"github.com/target/goalert/engine/verifymanager"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// Engine handles automatic escalation of unacknowledged(triggered) alerts, as well as
// passing to-be-sent notifications to the notification.Sender.
//
// Care is taken to ensure only one attempt is made per contact-method
// at a time, regardless of how many instances of the application may be running.
type Engine struct {
	b   *backend
	mgr *lifecycle.Manager

	*cycleMonitor

	shutdownCh  chan struct{}
	triggerCh   chan struct{}
	runLoopExit chan struct{}

	modules []processinglock.Module
	msg     *message.DB

	a   *alert.Store
	cfg *Config

	// needed to re-start river during switchover
	runCtx context.Context

	triggerPauseCh chan *pauseReq
}

var _ notification.ResultReceiver = &Engine{}

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
		cycleMonitor:   newCycleMonitor(),

		a: c.AlertStore,
	}

	p.mgr = lifecycle.NewManager(p._run, p._shutdown)
	err = p.mgr.SetPauseResumer(lifecycle.PauseResumerFunc(
		p._pause,
		p._resume,
	))
	if err != nil {
		return nil, err
	}

	rotMgr, err := rotationmanager.NewDB(ctx, db, c.RiverDBSQL)
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
	ncMgr, err := npcyclemanager.NewDB(ctx, db, c.AlertLogStore)
	if err != nil {
		return nil, errors.Wrap(err, "notification cycle backend")
	}
	statMgr, err := statusmgr.NewDB(ctx, db, c.DestRegistry, c.ConfigSource)
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
	cleanMgr, err := cleanupmanager.NewDB(ctx, db, c.AlertStore, c.Logger.With(slog.String("module", "cleanup")))
	if err != nil {
		return nil, errors.Wrap(err, "cleanup backend")
	}
	metricsMgr, err := metricsmanager.NewDB(ctx, db)
	if err != nil {
		return nil, errors.Wrap(err, "metrics management backend")
	}
	compatMgr, err := compatmanager.NewDB(ctx, db, c.SlackStore)
	if err != nil {
		return nil, errors.Wrap(err, "compatibility backend")
	}

	p.modules = []processinglock.Module{
		compatMgr,
		rotMgr,
		schedMgr,
		epMgr,
		ncMgr,
		statMgr,
		verifyMgr,
		hbMgr,
		cleanMgr,
		metricsMgr,
	}

	if expflag.ContextHas(ctx, expflag.UnivKeys) {
		signalMgr, err := signalmgr.NewDB(ctx, db)
		if err != nil {
			return nil, errors.Wrap(err, "signal manager backend")
		}
		p.modules = append(p.modules, signalMgr)
	}

	p.msg, err = message.NewDB(ctx, db, c.AlertLogStore, p.mgr)
	if err != nil {
		return nil, errors.Wrap(err, "messaging backend")
	}

	p.b, err = newBackend(db)
	if err != nil {
		return nil, errors.Wrap(err, "init backend")
	}

	args := processinglock.SetupArgs{
		DB:       db,
		Workers:  c.RiverWorkers,
		EventBus: c.EventBus,
		River:    c.River,
	}
	for _, m := range p.modules {
		if s, ok := m.(processinglock.Setupable); ok {
			err = s.Setup(ctx, args)
			if err != nil {
				return nil, errors.Wrap(err, "setup module")
			}
		}
	}

	return p, nil
}

func (p *Engine) AuthLinkURL(ctx context.Context, providerID, subjectID string, meta authlink.Metadata) (url string, err error) {
	permission.SudoContext(ctx, func(ctx context.Context) {
		url, err = p.cfg.AuthLinkStore.AuthLinkURL(ctx, providerID, subjectID, meta)
	})
	return url, err
}

func (p *Engine) processModule(ctx context.Context, m processinglock.Updatable) {
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
		if err != nil && !errors.Is(err, processinglock.ErrNoLock) {
			log.Log(ctx, errors.Wrap(err, m.Name()))
		}
		break
	}
}

func (p *Engine) processMessages(ctx context.Context) {
	defer recoverPanic(ctx, "MessageManager")
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	err := p.msg.SendMessages(ctx, p.sendMessage, p.cfg.NotificationManager.MessageStatus)
	if errors.Is(err, processinglock.ErrNoLock) {
		return
	}
	if errors.Is(err, message.ErrAbort) {
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
func (p *Engine) Trigger() { <-p.triggerCh }

// Pause will attempt to gracefully stop engine processing.
func (p *Engine) Pause(ctx context.Context) error {
	return p.mgr.Pause(ctx)
}

func (p *Engine) _pause(ctx context.Context) error {
	ch := make(chan error, 1)

	err := p.cfg.River.Stop(ctx)
	if err != nil {
		return errors.Wrap(err, "pause river")
	}

	select {
	case <-p.shutdownCh:
		return errors.New("shutting down")
	case <-ctx.Done():
		return ctx.Err()
	case p.triggerPauseCh <- &pauseReq{ch: ch, ctx: ctx}:
		select {
		case <-ctx.Done():
			defer func() { _ = p.Resume(ctx) }()
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
	if p.cfg.DisableCycle {
		return nil
	}

	err := p.cfg.River.QueueResume(ctx, "*", nil)
	if err != nil {
		return errors.Wrap(err, "resume river")
	}

	go p.startRiver()

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

	return p.mgr.Shutdown(ctx)
}

func (p *Engine) _shutdown(ctx context.Context) (err error) {
	close(p.shutdownCh)
	<-p.runLoopExit
	if !p.cfg.DisableCycle {
		err = p.cfg.River.Stop(ctx)
	}
	return err
}

// SetSendResult will update the status of a message.
func (p *Engine) SetSendResult(ctx context.Context, res *notification.SendResult) error {
	var err error
	permission.SudoContext(ctx, func(ctx context.Context) {
		err = p.msg.UpdateMessageStatus(ctx, res)
	})
	return err
}

// ReceiveSubject will process a notification result.
func (p *Engine) ReceiveSubject(ctx context.Context, providerID, subjectID, callbackID string, result notification.Result) error {
	cb, err := p.b.FindOne(ctx, callbackID)
	if err != nil {
		return err
	}
	if cb.ServiceID != "" {
		ctx = log.WithField(ctx, "ServiceID", cb.ServiceID)
	}
	if cb.AlertID != 0 {
		ctx = log.WithField(ctx, "AlertID", cb.AlertID)
	}

	var usr *user.User
	permission.SudoContext(ctx, func(ctx context.Context) {
		usr, err = p.cfg.UserStore.FindOneBySubject(ctx, providerID, subjectID)
	})
	if err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}
	if usr == nil {
		return &notification.UnknownSubjectError{
			AlertID: cb.AlertID,
		}
	}

	ctx = permission.UserSourceContext(ctx, usr.ID, usr.Role, &permission.SourceInfo{
		Type: permission.SourceTypeNotificationCallback,
		ID:   callbackID,
	})

	var newStatus alert.Status
	switch result {
	case notification.ResultAcknowledge:
		newStatus = alert.StatusActive
	case notification.ResultResolve:
		newStatus = alert.StatusClosed
	case notification.ResultEscalate:
		err = p.a.EscalateAsOf(ctx, cb.AlertID, cb.CreatedAt)
		if err != nil {
			return fmt.Errorf("escalate alert: %w", err)
		}
		return nil
	default:
		return errors.New("unknown result type")
	}

	if cb.AlertID != 0 {
		return errors.Wrap(p.a.UpdateStatus(ctx, cb.AlertID, newStatus), "update alert")
	}
	if cb.ServiceID != "" {
		return errors.Wrap(p.a.UpdateStatusByService(ctx, cb.ServiceID, newStatus), "update all alerts")
	}

	return errors.New("unknown callback type")
}

// Receive will process a notification result.
func (p *Engine) Receive(ctx context.Context, callbackID string, result notification.Result) error {
	cb, err := p.b.FindOne(ctx, callbackID)
	if err != nil {
		return err
	}
	if cb.ServiceID != "" {
		ctx = log.WithField(ctx, "ServiceID", cb.ServiceID)
	}
	if cb.AlertID != 0 {
		ctx = log.WithField(ctx, "AlertID", cb.AlertID)
	}

	var usr *user.User
	permission.SudoContext(ctx, func(ctx context.Context) {
		cm, serr := p.cfg.ContactMethodStore.FindOne(ctx, p.b.db, cb.ContactMethodID)
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

	var newStatus alert.Status
	switch result {
	case notification.ResultAcknowledge:
		newStatus = alert.StatusActive
	case notification.ResultResolve:
		newStatus = alert.StatusClosed
	case notification.ResultEscalate:
		err = p.a.EscalateAsOf(ctx, cb.AlertID, cb.CreatedAt)
		if err != nil {
			return fmt.Errorf("escalate alert: %w", err)
		}
		return nil
	default:
		return errors.New("unknown result type")
	}

	if cb.AlertID != 0 {
		return errors.Wrap(p.a.UpdateStatus(ctx, cb.AlertID, newStatus), "update alert")
	}
	if cb.ServiceID != "" {
		return errors.Wrap(p.a.UpdateStatusByService(ctx, cb.ServiceID, newStatus), "update all alerts")
	}

	return errors.New("unknown callback type")
}

// Start will enable all associated contact methods of `value` with type `t`. This should
// be invoked if a user, for example, responds with `START` via sms.
func (p *Engine) Start(ctx context.Context, d gadb.DestV1) error {
	var err error
	permission.SudoContext(ctx, func(ctx context.Context) {
		err = p.cfg.ContactMethodStore.EnableByDest(ctx, p.b.db, d)
	})

	if errors.Is(err, sql.ErrNoRows) {
		// no contact methods found, nothing to do
		return nil
	}

	return err
}

// Stop will disable all associated contact methods of `value` with type `t`. This should
// be invoked if a user, for example, responds with `STOP` via SMS.
func (p *Engine) Stop(ctx context.Context, d gadb.DestV1) error {
	var err error
	permission.SudoContext(ctx, func(ctx context.Context) {
		err = p.cfg.ContactMethodStore.DisableByDest(ctx, p.b.db, d)
	})

	if errors.Is(err, sql.ErrNoRows) {
		// no contact methods found, nothing to do
		return nil
	}

	return err
}

func (p *Engine) processAll(ctx context.Context) bool {
	for _, m := range p.modules {
		if p.mgr.IsPausing() {
			return true
		}

		up, ok := m.(processinglock.Updatable)
		if !ok {
			continue
		}

		start := time.Now()
		p.processModule(ctx, up)
		metricModuleDuration.WithLabelValues(m.Name()).Observe(time.Since(start).Seconds())
	}
	return false
}

func monitorCycle(ctx context.Context, start time.Time) (cancel func()) {
	ctx, cancel = context.WithCancel(ctx)

	go func() {
		watch := time.NewTicker(5 * time.Second)
		defer watch.Stop()
		watchErr := time.NewTicker(time.Minute)
		defer watchErr.Stop()

	loop:
		for {
			select {
			case <-watchErr.C:
				log.Log(log.WithField(ctx, "elapsedSec", time.Since(start).Seconds()), fmt.Errorf("engine possibly stuck"))
			case <-watch.C:
				log.Logf(log.WithField(ctx, "elapsedSec", time.Since(start).Seconds()), "long engine cycle")
			case <-ctx.Done():
				break loop
			}
		}

		dur := time.Since(start)
		if dur < 5*time.Second {
			return
		}

		log.Log(log.WithField(ctx, "elapsedSec", dur.Seconds()), fmt.Errorf("slow cycle finished"))
	}()

	return cancel
}

func (p *Engine) cycle(ctx context.Context) {
	// track start of next cycle, and defer the call to the returned sfinish function
	defer p.startNextCycle()()
	ctx = p.cfg.ConfigSource.Config().Context(ctx)

	if p.mgr.IsPausing() {
		log.Logf(ctx, "Engine cycle disabled (paused or shutting down).")
		return
	}

	if p.cfg.LogCycles {
		log.Logf(ctx, "Engine cycle start.")
		defer log.Logf(ctx, "Engine cycle end.")
	}

	startAll := time.Now()
	defer monitorCycle(ctx, startAll)()

	aborted := p.processAll(ctx)
	if aborted || p.mgr.IsPausing() {
		log.Logf(ctx, "Engine cycle aborted (paused or shutting down).")
		return
	}
	startMsg := time.Now()
	p.processMessages(ctx)
	metricModuleDuration.WithLabelValues("Engine.Message").Observe(time.Since(startMsg).Seconds())
	metricModuleDuration.WithLabelValues("Engine").Observe(time.Since(startAll).Seconds())
	metricCycleTotal.Inc()
}

func (p *Engine) handlePause(ctx context.Context, respCh chan error) {
	// nothing special to do currently
	respCh <- nil
}

func (p *Engine) startRiver() {
	ctx := p.runCtx
	err := p.cfg.River.Start(ctx)
	// Ignore shutdown errors, they are expected.
	// TODO: replace with errors.Is(err, river.ErrShutdown) when/if
	// this gets merged: https://github.com/riverqueue/river/pull/830
	if err != nil && !strings.Contains(err.Error(), "shutdown") {
		log.Log(ctx, errors.Wrap(err, "start river"))
	}
}

func (p *Engine) _run(ctx context.Context) error {
	defer close(p.runLoopExit)
	ctx = permission.SystemContext(ctx, "Engine")
	if p.cfg.DisableCycle {
		log.Logf(ctx, "Engine started in API-only mode.")
		ch := make(chan struct{})
		close(ch)
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

	p.runCtx = ctx
	go p.startRiver()

	dur := p.cfg.CycleTime
	if dur == 0 {
		dur = 5 * time.Second
	}
	alertTicker := time.NewTicker(dur)
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
