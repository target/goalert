package engine

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/target/goalert/engine/signal"
	"github.com/target/goalert/notification"
)

func (p *Engine) sendSignal(ctx context.Context, sig *signal.OutgoingSignal) (*notification.SendResult, error) {
	name, _, err := p.a.ServiceInfo(ctx, sig.ServiceID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup service info")
	}

	var notifMsg notification.Message

	switch sig.Dest.Type {
	case notification.DestTypeSlackChannel:
		notifMsg = notification.Signal{
			Dest:        sig.Dest,
			CallbackID:  sig.ID,
			SignalID:    sig.SignalID,
			Summary:     sig.Message,
			ServiceID:   sig.ServiceID,
			ServiceName: name,
		}
	case notification.DestTypeUserWebhook:
		notifMsg = notification.Signal{
			Dest:        sig.Dest,
			CallbackID:  sig.ID,
			SignalID:    sig.SignalID,
			Summary:     sig.Message,
			ServiceID:   sig.ServiceID,
			ServiceName: name,
		}
	case notification.DestTypeUserEmail:
		email := notification.SignalEmail{}
		err := json.Unmarshal(sig.Content, &email)
		if err != nil {
			return nil, err
		}

		notifMsg = notification.Signal{
			Dest:        sig.Dest,
			CallbackID:  sig.ID,
			SignalID:    sig.SignalID,
			Summary:     sig.Message,
			ServiceID:   sig.ServiceID,
			ServiceName: name,
			Email:       &email,
		}
	}

	res, err := p.cfg.NotificationManager.SendMessage(ctx, notifMsg)
	if err != nil {
		return nil, err
	}

	return res, nil
}
