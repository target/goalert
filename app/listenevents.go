package app

import (
	"context"
	"time"

	"github.com/jackc/pgx"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

func (app *App) listenEvents(ctx context.Context, db *sql.DB) error {
	l, err := sqlutil.NewListener(ctx, db, "/goalert/config-refresh")
	if err != nil {
		return err
	}
	app.events = l

	go func() {
		for {
			var n *pgx.Notification
			select {
			case n = <-l.NotificationChannel():
			case <-ctx.Done():
				return
			}
			if n == nil {
				return
			}

			log.Debugf(log.WithFields(ctx, log.Fields{
				"Channel": n.Channel,
				"PID":     n.BePid,
				"Extra":   n.Extra,
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
