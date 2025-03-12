package app

import (
	"context"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

func (app *App) setupListenEvents() {
	app.events = sqlutil.NewListener(app.pgx)
	app.events.Handle("/goalert/config-refresh", func(ctx context.Context, payload string) error {
		permission.SudoContext(ctx, func(ctx context.Context) {
			log.Log(ctx, app.ConfigStore.Reload(ctx))
		})
		return nil
	})
}
