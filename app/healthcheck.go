package app

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/target/goalert/app/lifecycle"
	"github.com/target/goalert/util/errutil"
)

func (app *App) healthCheck(w http.ResponseWriter, req *http.Request) {
	if app.mgr.Status() == lifecycle.StatusShutdown {
		http.Error(w, "server shutting down", http.StatusInternalServerError)
		return
	}
	if app.mgr.Status() == lifecycle.StatusStarting {
		http.Error(w, "server starting", http.StatusInternalServerError)
		return
	}

	// Good to go
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

	err := app.Engine.WaitNextCycle(req.Context())
	errutil.HTTPError(req.Context(), w, errors.Wrap(err, "engine cycle"))
}
