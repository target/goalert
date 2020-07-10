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

	if app.srv != nil {
		errs = append(errs, errors.Wrap(app.srv.Shutdown(ctx), "shutdown HTTP server"))
	}

	if app.Engine != nil {
		errs = append(errs, errors.Wrap(app.Engine.Shutdown(ctx), "shutdown engine"))
	}

	if app.events != nil {
		errs = append(errs, errors.Wrap(app.events.Close(), "close event listener"))
	}

	if app.notificationManager != nil {
		errs = append(errs, errors.Wrap(app.notificationManager.Shutdown(ctx), "shutdown notification manager"))
	}

	if app.SessionKeyring != nil {
		errs = append(errs, errors.Wrap(app.SessionKeyring.Shutdown(ctx), "shutdown session keyring"))
	}

	if app.OAuthKeyring != nil {
		errs = append(errs, errors.Wrap(app.OAuthKeyring.Shutdown(ctx), "shutdown oauth keyring"))
	}

	if app.NonceStore != nil {
		errs = append(errs, errors.Wrap(app.NonceStore.Shutdown(ctx), "shutdown nonce store"))
	}

	// filter out nil values
	shutdownErrs := errs[:0]
	for _, e := range errs {
		if e == nil {
			continue
		}
		shutdownErrs = append(shutdownErrs, e)
	}

	if len(shutdownErrs) == 1 {
		return shutdownErrs[0]
	}
	if len(shutdownErrs) > 1 {
		return errors.Errorf("multiple shutdown errors: %+v", shutdownErrs)
	}

	return nil
}

var shutdownSignals = []os.Signal{os.Interrupt}

const shutdownTimeout = time.Minute * 2
