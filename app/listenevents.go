package app

import (
	"context"

	"github.com/jackc/pgx"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

func (app *App) listenEvents(ctx context.Context) error {
	l, err := sqlutil.NewListener(ctx, (*sqlutil.DBConnector)(app.db), "/goalert/config-refresh")
	if err != nil {
		return err
	}
	app.events = l
	go func() {
		for err := range l.Errors() {
			log.Log(ctx, err)
		}
	}()

	go func() {
		for {
			var n *pgx.Notification
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

	return nil
}
