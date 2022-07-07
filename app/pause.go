package app

import (
	"context"

	"github.com/target/goalert/util/log"
)

// LogBackgroundContext returns a context.Background with the application logger configured.
func (app *App) LogBackgroundContext() context.Context { return app.cfg.Logger.BackgroundContext() }

func (app *App) Pause(ctx context.Context) error {
	ctx = log.WithLogger(ctx, app.cfg.Logger)

	err := app.mgr.Pause(ctx)
	if err != nil {
		return err
	}
	app.db.SetMaxIdleConns(0)
	return nil
}

func (app *App) Resume() {
	app.db.SetMaxIdleConns(app.cfg.DBMaxIdle)
	app.mgr.Resume(app.LogBackgroundContext())
}

func (app *App) _pause(ctx context.Context) error {
	app.db.SetMaxIdleConns(0)
	app.events.Stop()

	return nil
}

func (app *App) _resume(ctx context.Context) error {
	app.db.SetMaxIdleConns(app.cfg.DBMaxIdle)
	app.events.Start()

	return nil
}
