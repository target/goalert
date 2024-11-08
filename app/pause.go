package app

import (
	"context"
	"time"
)

// LogBackgroundContext returns a context.Background with the application logger configured.
func (app *App) LogBackgroundContext() context.Context {
	return app.cfg.LegacyLogger.BackgroundContext()
}

func (app *App) Pause(ctx context.Context) error {
	return app.mgr.Pause(app.Context(ctx))
}

func (app *App) Resume(ctx context.Context) error {
	return app.mgr.Resume(app.Context(ctx))
}

func (app *App) _pause(ctx context.Context) error {
	app.db.SetMaxIdleConns(0)
	app.db.SetConnMaxLifetime(time.Second)
	app.db.SetMaxOpenConns(3)
	app.events.Stop()
	return nil
}

func (app *App) _resume(ctx context.Context) error {
	app.db.SetMaxOpenConns(app.cfg.DBMaxOpen)
	app.db.SetMaxIdleConns(app.cfg.DBMaxIdle)
	app.db.SetConnMaxLifetime(0)
	app.events.Start()

	return nil
}
