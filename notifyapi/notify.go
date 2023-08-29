package notifyapi

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/validation/validate"
)

// Handler responds to generic API requests
type Handler struct {
	c Config
}

// NewHandler creates a new Handler, registering generic API routes using chi.
func NewHandler(c Config) *Handler {
	return &Handler{c: c}
}

// ServeCreateAlert allows creating or closing an alert.
func (h *Handler) ServeCreateAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := permission.LimitCheckAny(ctx, permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	serviceID := permission.ServiceID(ctx)

	summary := r.FormValue("summary")
	details := r.FormValue("details")
	action := r.FormValue("action")
	dedup := r.FormValue("dedup")

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct == "application/json" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var b struct {
			Summary, Details, Action, Dedup *string
		}
		err = json.Unmarshal(data, &b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if b.Summary != nil {
			summary = *b.Summary
		}
		if b.Details != nil {
			details = *b.Details
		}
		if b.Dedup != nil {
			dedup = *b.Dedup
		}
		if b.Action != nil {
			action = *b.Action
		}
	}

	status := alert.StatusTriggered
	if action == "close" {
		status = alert.StatusClosed
	}

	summary = validate.SanitizeText(summary, alert.MaxSummaryLength)
	details = validate.SanitizeText(details, alert.MaxDetailsLength)

	a := &alert.Alert{
		Summary:   summary,
		Details:   details,
		Source:    alert.SourceNotify,
		ServiceID: serviceID,
		Dedup:     alert.NewUserDedup(dedup),
		Status:    status,
	}

	var resp struct {
		AlertID   int
		ServiceID string
		IsNew     bool
	}

	err = retry.DoTemporaryError(func(int) error {
		createdAlert, isNew, err := h.c.AlertStore.CreateOrUpdate(ctx, a)
		if createdAlert != nil {
			resp.AlertID = createdAlert.ID
			resp.ServiceID = createdAlert.ServiceID
			resp.IsNew = isNew
		}

		return err
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second),
	)
	if errutil.HTTPError(ctx, w, errors.Wrap(err, "create alert")) {
		return
	}

	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(204)
		return
	}

	data, err := json.Marshal(&resp)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
