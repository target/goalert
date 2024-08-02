package notification

import (
	"context"
	"fmt"
	"sync"

	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/util/log"
)

// Manager is used as an intermediary between Senders and Receivers.
// It should be constructed first (with NewManager()) and passed to
// Senders and Receivers that require it.
type Manager struct {
	ResultReceiver
	mx *sync.RWMutex

	reg *nfydest.Registry
}

var _ ResultReceiver = Manager{}

// NewManager initializes a new Manager.
func NewManager(reg *nfydest.Registry) *Manager {
	return &Manager{
		mx:  new(sync.RWMutex),
		reg: reg,
	}
}

// MessageStatus will return the current status of a message.
func (mgr *Manager) MessageStatus(ctx context.Context, providerMsgID ProviderMessageID) (*Status, error) {
	return mgr.reg.MessageStatus(ctx, providerMsgID.ProviderName, providerMsgID.ExternalID)
}

// SetResultReceiver will set the ResultReceiver as the target for all Receiver calls.
// It will panic if called multiple times.
func (mgr *Manager) SetResultReceiver(ctx context.Context, recv ResultReceiver) error {
	if mgr.ResultReceiver != nil {
		panic("tried to register a second Processor instance")
	}
	mgr.ResultReceiver = recv

	types, err := mgr.reg.Types(ctx)
	if err != nil {
		return fmt.Errorf("error getting types: %w", err)
	}

	for _, t := range types {
		p := mgr.reg.Provider(t.Type)
		if r, ok := p.(ReceiverSetter); ok {
			r.SetReceiver(&namedReceiver{r: recv, destType: t.Type})
		}
	}

	return nil
}

// SendMessage tries all registered senders for the type given
// in Notification. An error is returned if there are no registered senders for the type
// or if an error is returned from all of them.
func (mgr *Manager) SendMessage(ctx context.Context, msg Message) (*SendResult, error) {
	mgr.mx.RLock()
	defer mgr.mx.RUnlock()

	destType := msg.DestType()

	ctx = log.WithFields(ctx, log.Fields{
		"ProviderType": destType,
		"CallbackID":   msg.MsgID(),
	})
	if a, ok := msg.(Alert); ok {
		ctx = log.WithField(ctx, "AlertID", a.AlertID)
	}
	sendCtx := log.WithField(ctx, "ProviderName", msg.DestType())
	sent, err := mgr.reg.SendMessage(sendCtx, msg)
	if err != nil {
		return nil, err
	}

	log.Logf(sendCtx, "notification sent")
	metricSentTotal.
		WithLabelValues(msg.DestType(), fmt.Sprintf("%T", msg), msgSvcID(msg)).
		Inc()

	return &SendResult{
		ID: msg.MsgID(),
		Status: Status{
			State:    sent.State,
			Details:  sent.StateDetails,
			SrcValue: sent.SrcValue,
		},
		ProviderMessageID: ProviderMessageID{
			ProviderName: msg.DestType(),
			ExternalID:   sent.ExternalID,
		},
	}, nil
}

func msgSvcID(msg Message) string {
	switch msg := msg.(type) {
	case Alert:
		return msg.ServiceID
	case AlertBundle:
		return msg.ServiceID
	case AlertStatus:
		return msg.ServiceID
	}

	return ""
}
