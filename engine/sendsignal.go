package engine

import (
	"context"

	"github.com/pkg/errors"
	"github.com/target/goalert/engine/signal"
	"github.com/target/goalert/notification"
)

func (p *Engine) sendSignal(ctx context.Context, sig *signal.OutgoingSignal) (*notification.SendResult, error) {
	name, _, err := p.a.ServiceInfo(ctx, sig.ServiceID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup service info")
	}

	notifMsg := notification.Signal{
		Dest:        sig.Dest,
		CallbackID:  sig.ID,
		SignalID:    sig.SignalID,
		Summary:     sig.Message,
		ServiceID:   sig.ServiceID,
		ServiceName: name,
	}

	res, err := p.cfg.NotificationManager.SendMessage(ctx, notifMsg)
	if err != nil {
		return nil, err
	}

	return res, nil
}
