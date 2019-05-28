package app

import (
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

func (app *App) healthCheck(w http.ResponseWriter, req *http.Request) {
	if app.mgr.Status() == lifecycle.StatusShutdown {
		http.Error(w, "server shutting down", http.StatusInternalServerError)
		return
	}

	ctx := req.Context()
	err := retry.DoTemporaryError(func(_ int) error {
		return app.db.PingContext(ctx)
	},
		retry.Log(ctx),
		retry.Limit(5),
		retry.FibBackoff(100*time.Millisecond),
	)

	errutil.HTTPError(req.Context(), w, errors.Wrap(err, "engine cycle"))
}

func (app *App) engineStatus(w http.ResponseWriter, req *http.Request) {
	if app.mgr.Status() == lifecycle.StatusShutdown {
		http.Error(w, "server shutting down", http.StatusInternalServerError)
		return
	}

	if app.cfg.APIOnly {
		http.Error(w, "engine not running", http.StatusInternalServerError)
		return
	}

	err := app.engine.WaitNextCycle(req.Context())
	errutil.HTTPError(req.Context(), w, errors.Wrap(err, "engine cycle"))
}
