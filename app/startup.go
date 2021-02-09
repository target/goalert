package app

import (
	"context"
	"time"

	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/email"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
)

func (app *App) initStartup(ctx context.Context, label string, fn func(context.Context) error) {
	if app.startupErr != nil {
		return
	}

	ctx, sp := trace.StartSpan(ctx, label)
	defer sp.End()
	err := fn(ctx)
	if err != nil {
		sp.Annotate([]trace.Attribute{trace.BoolAttribute("error", true)}, err.Error())
		app.startupErr = errors.Wrap(err, label)
	}
}

func (app *App) startup(ctx context.Context) error {
	ctx, sp := trace.StartSpan(ctx, "Startup")
	defer sp.End()

	app.initStartup(ctx, "Startup.TestDBConn", func(ctx context.Context) error {
		err := app.db.PingContext(ctx)
		if err == nil {
			return nil
		}

		t := time.NewTicker(time.Second)
		defer t.Stop()
		for retry.IsTemporaryError(err) {
			log.Log(ctx, err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				err = app.db.PingContext(ctx)
			}
		}

		return err
	})

	app.notificationManager = notification.NewManager()
	if app.cfg.StubNotifiers {
		app.notificationManager.SetStubNotifiers()
	}

	app.initStartup(ctx, "Startup.DBStores", app.initStores)

	// init twilio before engine
	app.initStartup(
		ctx, "Startup.Twilio", app.initTwilio)

	app.initStartup(ctx, "Startup.Slack", app.initSlack)
	app.notificationManager.RegisterSender(notification.DestTypeUserEmail, "smtp", email.NewSender(ctx))
	app.notificationManager.RegisterSender(notification.DestTypeUserWebhook, "webhook", webhook.NewSender(ctx))

	app.initStartup(ctx, "Startup.Engine", app.initEngine)
	app.initStartup(ctx, "Startup.Auth", app.initAuth)
	app.initStartup(ctx, "Startup.GraphQL", app.initGraphQL)

	app.initStartup(ctx, "Startup.HTTPServer", app.initHTTP)

	if app.startupErr != nil {
		return app.startupErr
	}

	return app.mgr.SetPauseResumer(lifecycle.MultiPauseResume(
		app.Engine,
		lifecycle.PauseResumerFunc(app._pause, app._resume),
	))
}
