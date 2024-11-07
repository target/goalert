package app

import (
	"context"
	"net/http"
	"os"

	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

var triggerSignals []os.Signal

// Run will start the application and start serving traffic.
func (app *App) Run(ctx context.Context) error {
	return app.mgr.Run(app.Context(ctx))
}

func (app *App) _Run(ctx context.Context) error {
	go func() {
		err := app.Engine.Run(ctx)
		if err != nil {
			log.Log(ctx, err)
		}
	}()

	go func() {
		err := app.RiverUI.Start(ctx)
		if err != nil {
			log.Log(ctx, err)
		}
	}()

	eventCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	eventDoneCh, err := app.listenEvents(eventCtx)
	if err != nil {
		return err
	}

	if app.sysAPISrv != nil {
		log.Logf(log.WithField(ctx, "address", app.sysAPIL.Addr().String()), "System API server started.")
		go func() {
			if err := app.sysAPISrv.Serve(app.sysAPIL); err != nil {
				log.Log(ctx, err)
			}
		}()
	}

	if app.smtpsrv != nil {
		log.Logf(log.WithField(ctx, "address", app.smtpsrvL.Addr().String()), "SMTP server started.")
		go func() {
			if err := app.smtpsrv.ServeSMTP(app.smtpsrvL); err != nil {
				log.Log(ctx, err)
			}
		}()
	}

	log.Logf(
		log.WithFields(ctx, log.Fields{
			"address": app.l.Addr().String(),
			"url":     app.ConfigStore.Config().PublicURL(),
		}),
		"Listening.",
	)
	err = app.srv.Serve(app.l)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "serve HTTP")
	}
	if app.hSrv != nil {
		app.hSrv.Resume()
	}

	select {
	case <-eventDoneCh:
	case <-ctx.Done():
	}
	return nil
}
