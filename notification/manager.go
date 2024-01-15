package notification

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
)

// Manager is used as an intermediary between Senders and Receivers.
// It should be constructed first (with NewManager()) and passed to
// Senders and Receivers that require it.
type Manager struct {
	providers   map[string]*namedSender
	searchOrder []*namedSender

	ResultReceiver
	mx *sync.RWMutex

	stubNotifiers bool
}

var _ ResultReceiver = Manager{}

// NewManager initializes a new Manager.
func NewManager() *Manager {
	return &Manager{
		mx:        new(sync.RWMutex),
		providers: make(map[string]*namedSender),
	}
}

// SetStubNotifiers will cause all notifications senders to be stubbed out.
//
// This causes all notifications to be marked as delivered, but not actually sent.
func (mgr *Manager) SetStubNotifiers() {
	mgr.stubNotifiers = true
}

// FormatDestValue will format the destination value if an available FriendlyValuer exists
// for the destType or return the original.
func (mgr *Manager) FormatDestValue(ctx context.Context, destType DestType, value string) string {
	if value == "" {
		return ""
	}
	mgr.mx.RLock()
	defer mgr.mx.RUnlock()

	for _, s := range mgr.searchOrder {
		if s.destType != destType {
			continue
		}

		f, ok := s.Sender.(FriendlyValuer)
		if !ok {
			continue
		}

		newValue, err := f.FriendlyValue(ctx, value)
		if err != nil {
			log.Log(ctx, fmt.Errorf("format dest value with '%s': %w", s.name, err))
			continue
		}

		return newValue
	}
	return value
}

// MessageStatus will return the current status of a message.
func (mgr *Manager) MessageStatus(ctx context.Context, providerMsgID ProviderMessageID) (*Status, DestType, error) {
	provider := mgr.providers[providerMsgID.ProviderName]
	if provider == nil {
		return nil, DestTypeUnknown, errors.Errorf("unknown provider ID '%s'", providerMsgID.ProviderName)
	}

	checker, ok := provider.Sender.(StatusChecker)
	if !ok {
		return nil, DestTypeUnknown, ErrStatusUnsupported
	}

	status, err := checker.Status(ctx, providerMsgID.ExternalID)
	return status, provider.destType, err
}

// RegisterSender will register a sender under a given DestType and name.
// A sender for the same name and type will replace an existing one, if any.
func (mgr *Manager) RegisterSender(t DestType, name string, s Sender) {
	mgr.mx.Lock()
	defer mgr.mx.Unlock()

	_, ok := mgr.providers[name]
	if ok {
		panic("name already taken")
	}
	if mgr.stubNotifiers {
		// disable notification sending
		s = stubSender{}
	}

	n := &namedSender{name: name, Sender: s, destType: t}
	mgr.providers[name] = n
	mgr.searchOrder = append(mgr.searchOrder, n)

	if rs, ok := s.(ReceiverSetter); ok {
		rs.SetReceiver(&namedReceiver{ns: n, r: mgr})
	}
}

// SetResultReceiver will set the ResultReceiver as the target for all Receiver calls.
// It will panic if called multiple times.
func (mgr *Manager) SetResultReceiver(p ResultReceiver) {
	if mgr.ResultReceiver != nil {
		panic("tried to register a second Processor instance")
	}
	mgr.ResultReceiver = p
}

// SendMessage tries all registered senders for the type given
// in Notification. An error is returned if there are no registered senders for the type
// or if an error is returned from all of them.
func (mgr *Manager) SendMessage(ctx context.Context, msg Message) (*SendResult, error) {
	mgr.mx.RLock()
	defer mgr.mx.RUnlock()

	destType := msg.Destination().Type

	ctx = log.WithFields(ctx, log.Fields{
		"ProviderType": destType,
		"CallbackID":   msg.ID(),
	})
	if a, ok := msg.(Alert); ok {
		ctx = log.WithField(ctx, "AlertID", a.AlertID)
	}

	var tried bool
	for _, s := range mgr.searchOrder {
		if s.destType != destType {
			continue
		}
		tried = true

		sendCtx := log.WithField(ctx, "ProviderName", s.name)
		res, err := s.Send(sendCtx, msg)
		if err != nil {
			log.Log(sendCtx, errors.Wrap(err, "send notification"))
			continue
		}
		log.Logf(sendCtx, "notification sent")
		metricSentTotal.
			WithLabelValues(msg.Destination().Type.String(), msg.Type().String(), msgSvcID(msg)).
			Inc()
		// status already wrapped via namedSender
		return res, nil
	}
	if !tried {
		return nil, fmt.Errorf("no senders registered for type '%s'", destType)
	}

	return nil, errors.New("all notification senders failed")
}

func msgSvcID(msg Message) string {
	switch msg := msg.(type) {
	case Alert:
		return msg.ServiceID
	case AlertBundle:
		return msg.ServiceID
	}

	return ""
}
