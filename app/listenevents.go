package app

import (
	"context"
	"time"

	"github.com/jackc/pgx"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

func (app *App) listenEvents(ctx context.Context) error {
	channels := []string{"/goalert/config-refresh"}

	handle := func(l *sqlutil.Listener) {
		defer l.Close()

		for {
			var n *pgx.Notification
			select {
			case n = <-l.NotificationChannel():
			case <-ctx.Done():
				return
			}

			if n == nil {
				continue
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
	}

	makeListener := func(url string) (*sqlutil.Listener, error) {
		l := sqlutil.NewListener(app.cfg.DBURL, 3*time.Second, time.Minute, nil)
		for _, ch := range channels {
			err := l.Listen(ch)
			if err != nil {
				l.Close()
				return nil, err
			}
		}
		err := l.Ping()
		if err != nil {
			l.Close()
			return nil, err
		}

		return l, nil
	}

	l, err := makeListener(app.cfg.DBURL)
	if err != nil {
		return err
	}
	var ln *sqlutil.Listener
	if app.cfg.DBURLNext != "" {
		ln, err = makeListener(app.cfg.DBURLNext)
		if err != nil {
			l.Close()
			return err
		}
	}

	go handle(l)
	if ln != nil {
		go handle(ln)
	}

	return nil
}
