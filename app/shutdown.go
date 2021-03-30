package app

import (
	"context"
	"os"
	"time"

	"github.com/pkg/errors"
)

// Shutdown will cause the App to begin a graceful shutdown, using
// the provided context for any cleanup operations.
func (app *App) Shutdown(ctx context.Context) error {
	return app.mgr.Shutdown(ctx)
}

func (app *App) _Shutdown(ctx context.Context) error {
	defer close(app.doneCh)
	defer app.db.Close()
	var errs []error

	if app.cooldown != nil {
		// wait for the cooldown (since last req closed)
		app.cooldown.WaitContext(ctx)
	}

	type shutdownable interface{ Shutdown(context.Context) error }

	shut := func(sh shutdownable, msg string) {
		if sh == nil {
			return
		}
		err := sh.Shutdown(ctx)
		if err != nil {
			errs = append(errs, errors.Wrap(err, msg))
		}
	}

	// It's important to shutdown the HTTP server first
	// so things like message responses are handled before
	// shutting down things like the engine or notification manager
	// that would still need to process them.
	shut(app.srv, "HTTP server")
	shut(app.Engine, "engine")
	shut(app.events, "event listener")
	shut(app.notificationManager, "notification manager")
	shut(app.SessionKeyring, "session keyring")
	shut(app.OAuthKeyring, "oauth keyring")
	shut(app.APIKeyring, "API keyring")
	shut(app.NonceStore, "nonce store")
	shut(app.ConfigStore, "config store")
	shut(app.requestLock, "context locker")

	if len(errs) == 1 {
		return errs[0]
	}
	if len(errs) > 1 {
		return errors.Errorf("multiple shutdown errors: %+v", errs)
	}

	return nil
}

var shutdownSignals = []os.Signal{os.Interrupt}

const shutdownTimeout = time.Minute * 2
