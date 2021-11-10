package app

import (
	"context"
	"net/http"

	"github.com/target/goalert/switchover"
	"github.com/target/goalert/util/log"

	"go.opencensus.io/trace"
)

func (app *App) pauseHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		err := app.requestLock.RLock(ctx)
		if err != nil {
			log.Log(ctx, err)
			return
		}
		defer app.requestLock.RUnlock()
		next.ServeHTTP(w, req)
	})
}

// LogContext returns a context.Background with the application logger configured.
func (app *App) LogContext() context.Context { return app.cfg.Logger.Context() }

func (app *App) Pause(ctx context.Context) error {
	ctx, sp := trace.StartSpan(ctx, "App.Pause")
	defer sp.End()

	err := app.mgr.Pause(ctx)
	if err != nil {
		return err
	}
	app.db.SetMaxIdleConns(0)
	return nil
}
func (app *App) Resume() {
	app.db.SetMaxIdleConns(app.cfg.DBMaxIdle)
	app.mgr.Resume(app.LogContext())
}
func (app *App) _pause(ctx context.Context) error {
	app.events.Stop()

	cfg := switchover.ConfigFromContext(ctx)
	if cfg.NoPauseAPI {
		return nil
	}
	err := app.requestLock.Lock(ctx)
	if err != nil {
		app.events.Start()
		return err
	}
	return nil
}
func (app *App) _resume(ctx context.Context) error {
	app.events.Start()
	app.requestLock.Unlock()
	return nil
}
