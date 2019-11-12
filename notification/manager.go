package notification

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
	"go.opencensus.io/trace"
)

// Manager is used as an intermediary between Senders and Receivers.
// It should be contstructed first (with NewManager()) and passed to
// Senders and Receivers that require it.
type Manager struct {
	providers   map[string]*namedSender
	searchOrder []*namedSender

	r  Receiver
	mx *sync.RWMutex

	shutdownCh chan struct{}
	shutdownWg sync.WaitGroup

	stubNotifiers bool
}

var _ Sender = &Manager{}

// NewManager initializes a new Manager.
func NewManager() *Manager {
	return &Manager{
		mx:         new(sync.RWMutex),
		shutdownCh: make(chan struct{}),
		providers:  make(map[string]*namedSender),
	}
}

// SetStubNotifiers will cause all notifications senders to be stubbed out.
//
// This causes all notifications to be marked as delivered, but not actually sent.
func (mgr *Manager) SetStubNotifiers() {
	mgr.stubNotifiers = true
}

func bgSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	var sp *trace.Span
	if ctx != nil {
		sp = trace.FromContext(ctx)
	}
	if sp == nil {
		return trace.StartSpan(context.Background(), name)
	}

	return trace.StartSpanWithRemoteParent(context.Background(), name, sp.SpanContext())
}

// Shutdown will stop the manager, waiting for pending background operations to finish.
func (m *Manager) Shutdown(context.Context) error {
	close(m.shutdownCh)
	m.shutdownWg.Wait()
	return nil
}

func (m *Manager) senderLoop(s *namedSender) {
	defer m.shutdownWg.Done()

	handleResponse := func(resp *MessageResponse) {
		ctx, sp := bgSpan(resp.Ctx, "NotificationManager.Response")
		cpy := *resp
		cpy.Ctx = ctx

		err := m.receive(ctx, s.name, &cpy)
		sp.End()
		resp.Err <- err
	}

	for {
		select {
		case resp := <-s.ListenResponse():
			handleResponse(resp)
		default:
		}

		select {
		case resp := <-s.ListenResponse():
			handleResponse(resp)
		case stat := <-s.ListenStatus():
			ctx, sp := bgSpan(stat.Ctx, "NotificationManager.StatusUpdate")
			m.updateStatus(ctx, stat.wrap(ctx, s))
			sp.End()
		case <-m.shutdownCh:
			return
		}
	}
}

// Status will return the current status of a message.
func (m *Manager) Status(ctx context.Context, id, providerMsgID string) (*MessageStatus, error) {
	parts := strings.SplitN(providerMsgID, ":", 2)
	if len(parts) != 2 {
		return nil, errors.Errorf("invalid provider message ID '%s'", providerMsgID)
	}

	provider := m.providers[parts[0]]
	if provider == nil {
		return nil, errors.Errorf("unknown provider ID '%s'", parts[0])
	}

	ctx, sp := trace.StartSpan(ctx, "NotificationManager.Status")
	sp.AddAttributes(
		trace.StringAttribute("provider.id", parts[0]),
		trace.StringAttribute("provider.message.id", parts[1]),
	)
	defer sp.End()
	stat, err := provider.Status(ctx, id, parts[1])
	if stat != nil {
		stat = stat.wrap(ctx, provider)
	}
	return stat, err
}

// RegisterSender will register a sender under a given DestType and name.
// A sender for the same name and type will replace an existing one, if any.
func (m *Manager) RegisterSender(t DestType, name string, s SendResponder) {
	m.mx.Lock()
	defer m.mx.Unlock()

	_, ok := m.providers[name]
	if ok {
		panic("name already taken")
	}
	if m.stubNotifiers {
		// disable notification sending
		s = stubSender{}
	}

	n := &namedSender{name: name, SendResponder: s, destType: t}
	m.providers[name] = n
	m.searchOrder = append(m.searchOrder, n)
	m.shutdownWg.Add(1)
	go m.senderLoop(n)
}

// UpdateStatus will update the status of a message.
func (m *Manager) updateStatus(ctx context.Context, status *MessageStatus) {
	err := m.r.UpdateStatus(ctx, status)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "update message status"))
	}
}

// RegisterReceiver will set the given Receiver as the target for all Receive() calls.
// It will panic if called multiple times.
func (m *Manager) RegisterReceiver(r Receiver) {
	if m.r != nil {
		panic("tried to register a second Receiver")
	}
	m.r = r
}

// Send implements the Sender interface by trying all registered senders for the type given
// in Notification. An error is returned if there are no registered senders for the type
// or if an error is returned from all of them.
func (m *Manager) Send(ctx context.Context, msg Message) (*MessageStatus, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	destType := msg.Destination().Type

	ctx = log.WithFields(ctx, log.Fields{
		"ProviderType": destType,
		"CallbackID":   msg.ID(),
	})
	if a, ok := msg.(Alert); ok {
		ctx = log.WithField(ctx, "AlertID", a.AlertID)
	}

	var tried bool
	for _, s := range m.searchOrder {
		if s.destType != destType {
			continue
		}
		tried = true

		sendCtx := log.WithField(ctx, "ProviderName", s.name)
		sendCtx, sp := trace.StartSpan(sendCtx, "NotificationManager.Send")
		sp.AddAttributes(
			trace.StringAttribute("provider.id", s.name),
			trace.StringAttribute("message.type", msg.Type().String()),
			trace.StringAttribute("message.id", msg.ID()),
		)
		status, err := s.Send(sendCtx, msg)
		sp.End()
		if err != nil {
			log.Log(sendCtx, errors.Wrap(err, "send notification"))
			continue
		}
		log.Debugf(sendCtx, "notification sent")
		// status already wrapped via namedSender
		return status, nil
	}
	if !tried {
		return nil, fmt.Errorf("no senders registered for type '%s'", destType)
	}

	return nil, errors.New("all notification senders failed")
}

func (m *Manager) receive(ctx context.Context, providerID string, resp *MessageResponse) error {
	ctx, sp := trace.StartSpan(ctx, "NotificationManager.Receive")
	defer sp.End()
	sp.AddAttributes(
		trace.StringAttribute("provider.id", providerID),
		trace.StringAttribute("message.id", resp.ID),
		trace.StringAttribute("dest.type", string(resp.From.Type)),
		trace.StringAttribute("dest.value", resp.From.Value),
		trace.StringAttribute("response", resp.Result.String()),
	)
	log.Debugf(log.WithFields(ctx, log.Fields{
		"Result":     resp.Result,
		"CallbackID": resp.ID,
		"ProviderID": providerID,
	}),
		"response received",
	)
	if resp.Result == ResultStop {
		return m.r.Stop(ctx, resp.From)
	}

	return m.r.Receive(ctx, resp.ID, resp.Result)
}
