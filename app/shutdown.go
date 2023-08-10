package app

import (
	"context"
	"os"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

// Shutdown will cause the App to begin a graceful shutdown, using
// the provided context for any cleanup operations.
func (app *App) Shutdown(ctx context.Context) error {
	return app.mgr.Shutdown(app.Context(ctx))
}

func (app *App) _Shutdown(ctx context.Context) error {
	defer close(app.doneCh)
	defer app.db.Close()
	var errs []error
	if app.hSrv != nil {
		app.hSrv.Shutdown()
	}

	type shutdownable interface{ Shutdown(context.Context) error }

	shut := func(sh shutdownable, msg string) {
		if sh == nil {
			return
		}
		t := reflect.TypeOf(sh)
		if reflect.ValueOf(sh) == reflect.Zero(t) {
			// check for nil pointer
			return
		}
		err := sh.Shutdown(ctx)
		if err != nil {
			errs = append(errs, errors.Wrap(err, msg))
		}
	}

	if app.sysAPISrv != nil {
		waitCh := make(chan struct{})
		go func() {
			defer close(waitCh)
			app.sysAPISrv.GracefulStop()
		}()
		select {
		case <-ctx.Done():
		case <-waitCh:
		}
		app.sysAPISrv.Stop()
	}

	// It's important to shutdown the HTTP server first
	// so things like message responses are handled before
	// shutting down things like the engine or notification manager
	// that would still need to process them.
	shut(app.smtpsrv, "SMTP receiver server")
	shut(app.srv, "HTTP server")
	shut(app.Engine, "engine")
	shut(app.events, "event listener")
	shut(app.SessionKeyring, "session keyring")
	shut(app.OAuthKeyring, "oauth keyring")
	shut(app.APIKeyring, "API keyring")
	shut(app.AuthLinkKeyring, "auth link keyring")
	shut(app.NonceStore, "nonce store")
	shut(app.ConfigStore, "config store")

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
