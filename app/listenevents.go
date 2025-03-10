package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

func (app *App) listenEvents(ctx context.Context) (<-chan struct{}, error) {
	l := app.events
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-l.Errors():
				log.Log(ctx, errors.Wrap(err, "listen events"))
			}
		}
	}()

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		for {
			var n *pgconn.Notification
			select {
			case n = <-l.Notifications():
			case <-ctx.Done():
				return
			}
			if n == nil {
				return
			}

			log.Debugf(log.WithFields(ctx, log.Fields{
				"Channel": n.Channel,
				"PID":     n.PID,
				"Payload": n.Payload,
			}), "NOTIFY")

			switch n.Channel {
			case "/goalert/config-refresh":
				permission.SudoContext(ctx, func(ctx context.Context) {
					log.Log(ctx, app.ConfigStore.Reload(ctx))
				})
			}
		}
	}()

	return doneCh, nil
}
