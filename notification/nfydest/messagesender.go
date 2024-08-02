package nfydest

import (
	"context"

	"github.com/target/goalert/notification/nfymsg"
)

// A MessageSender can send notifications.
type MessageSender interface {
	// Send should return nil error if the notification was sent successfully. It should be expected
	// that a returned error means that the notification should be attempted again.
	//
	// If the sent message can have its status tracked, a unique externalID should be returned.
	SendMessage(context.Context, nfymsg.Message) (*nfymsg.SentMessage, error)
}

func (r *Registry) SendMessage(ctx context.Context, msg nfymsg.Message) (*nfymsg.SentMessage, error) {
	p := r.Provider(msg.DestType())
	if p == nil {
		return nil, ErrUnknownType
	}

	s, ok := p.(MessageSender)
	if !ok {
		return nil, ErrUnsupported
	}

	info, err := p.TypeInfo(ctx)
	if err != nil {
		return nil, err
	}

	if !info.Enabled {
		return nil, ErrNotEnabled
	}

	switch msg.(type) {
	case nfymsg.Alert, nfymsg.AlertBundle:
		if !info.SupportsAlertNotifications {
			return nil, ErrUnsupported
		}
	case nfymsg.AlertStatus:
		if !info.SupportsStatusUpdates {
			return nil, ErrUnsupported
		}
	case nfymsg.Verification:
		if !info.SupportsUserVerification {
			return nil, ErrUnsupported
		}
	case nfymsg.Test:
	case nfymsg.SignalMessage:
		if !info.SupportsSignals {
			return nil, ErrUnsupported
		}
	case nfymsg.ScheduleOnCallUsers:
		if !info.SupportsOnCallNotify {
			return nil, ErrUnsupported
		}
	default:
		return nil, ErrUnsupported
	}

	if r.stubSender {
		return &nfymsg.SentMessage{
			ExternalID: "stub",
			State:      nfymsg.StateDelivered,
		}, nil
	}

	return s.SendMessage(ctx, msg)
}
