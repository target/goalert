package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/expflag"
	"github.com/target/goalert/notification/email"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/retry"

	"github.com/pkg/errors"
)

func (app *App) initStartup(ctx context.Context, label string, fn func(context.Context) error) {
	if app.startupErr != nil {
		return
	}

	err := fn(ctx)
	if err != nil {
		app.startupErr = errors.Wrap(err, label)
	}
}

func (app *App) startup(ctx context.Context) error {
	for _, f := range app.cfg.ExpFlags {
		if expflag.Description(f) == "" {
			app.Logger.WarnContext(ctx, "Unknown experimental flag.", slog.String("flag", string(f)))
		} else {
			app.Logger.InfoContext(ctx, "Experimental flag enabled.",
				slog.String("flag", string(f)),
				slog.String("description", expflag.Description(f)),
			)
		}
	}

	app.initStartup(ctx, "Startup.TestDBConn", func(ctx context.Context) error {
		err := app.db.PingContext(ctx)
		if err == nil { // success
			return nil
		}

		t := time.NewTicker(time.Second)
		defer t.Stop()
		for retry.IsTemporaryError(err) {
			app.Logger.WarnContext(ctx, "Failed to connect to database, will retry.", slog.Any("error", err))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				err = app.db.PingContext(ctx)
			}
		}

		return err
	})

	app.initStartup(ctx, "Startup.River", app.initRiver)
	app.initStartup(ctx, "Startup.DBStores", app.initStores)
	if app.startupErr != nil {
		return app.startupErr // ConfigStore will panic if not initialized
	}
	ctx = app.ConfigStore.Config().Context(ctx)

	// init twilio before engine
	app.initStartup(
		ctx, "Startup.Twilio", app.initTwilio)

	app.initStartup(ctx, "Startup.Slack", app.initSlack)

	app.initStartup(ctx, "Startup.Engine", app.initEngine)
	app.initStartup(ctx, "Startup.Auth", app.initAuth)
	app.initStartup(ctx, "Startup.GraphQL", app.initGraphQL)

	app.initStartup(ctx, "Startup.HTTPServer", app.initHTTP)
	app.initStartup(ctx, "Startup.SysAPI", app.initSysAPI)

	app.initStartup(ctx, "Startup.SMTPServer", app.initSMTPServer)

	if app.startupErr != nil {
		return app.startupErr
	}

	app.DestRegistry.RegisterProvider(ctx, app.twilioSMS)
	app.DestRegistry.RegisterProvider(ctx, app.twilioVoice)
	app.DestRegistry.RegisterProvider(ctx, email.NewSender(ctx))
	app.DestRegistry.RegisterProvider(ctx, app.ScheduleStore)
	app.DestRegistry.RegisterProvider(ctx, app.UserStore)
	app.DestRegistry.RegisterProvider(ctx, app.RotationStore)
	app.DestRegistry.RegisterProvider(ctx, app.AlertStore)
	app.DestRegistry.RegisterProvider(ctx, app.slackChan)
	app.DestRegistry.RegisterProvider(ctx, app.slackChan.DMSender())
	app.DestRegistry.RegisterProvider(ctx, app.slackChan.UserGroupSender())
	app.DestRegistry.RegisterProvider(ctx, webhook.NewSender(ctx, app.WebhookKeyring))
	if app.cfg.StubNotifiers {
		app.DestRegistry.StubNotifiers()
	}

	err := app.notificationManager.SetResultReceiver(ctx, app.Engine)
	if err != nil {
		return err
	}

	err = app.mgr.SetPauseResumer(lifecycle.MultiPauseResume(
		app.Engine,
		lifecycle.PauseResumerFunc(app._pause, app._resume),
	))
	if err != nil {
		return err
	}

	if app.cfg.SWO != nil {
		app.cfg.SWO.SetPauseResumer(app)
		app.Logger.InfoContext(ctx, "SWO Enabled.")
	}

	return nil
}
