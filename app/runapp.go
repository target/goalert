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
	return app.mgr.Run(ctx)
}

func (app *App) _Run(ctx context.Context) error {
	go func() {
		err := app.engine.Run(ctx)
		if err != nil {
			log.Log(ctx, err)
		}
	}()
	log.Logf(
		log.WithFields(context.TODO(), log.Fields{
			"address": app.l.Addr().String(),
			"url":     app.ConfigStore.Config().PublicURL(),
		}),
		"Listening.",
	)

	eventCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := app.listenEvents(eventCtx)
	if err != nil {
		return err
	}

	err = app.srv.Serve(app.l)
	if err != nil && err != http.ErrServerClosed {
		return errors.Wrap(err, "serve HTTP")
	}

	return nil
}
