package app

import (
	"io"
	"net/http"

	"github.com/google/uuid"
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

	var id uuid.UUID
	if nStr := req.FormValue("id"); nStr != "" {
		_id, err := uuid.Parse(nStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		id = _id
	} else {
		id = app.Engine.NextCycleID()
	}

	errutil.HTTPError(req.Context(), w, app.Engine.WaitCycleID(req.Context(), id))
}

func (app *App) engineCycle(w http.ResponseWriter, req *http.Request) {
	if app.mgr.Status() == lifecycle.StatusShutdown {
		http.Error(w, "server shutting down", http.StatusBadRequest)
		return
	}

	if app.cfg.APIOnly {
		http.Error(w, "engine not running", http.StatusBadRequest)
		return
	}

	io.WriteString(w, app.Engine.NextCycleID().String())
}
