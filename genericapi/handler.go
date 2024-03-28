package genericapi

import (
	"database/sql"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/auth"
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

// ServeUserAvatar will serve a redirect for a users avatar image.
func (h *Handler) ServeUserAvatar(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(req.URL.Path, "/")
	userID := parts[len(parts)-1]

	ctx := req.Context()
	u, err := h.c.UserStore.FindOne(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, req)
		return
	}
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	fullSize := req.FormValue("size") == "large"
	http.Redirect(w, req, u.ResolveAvatarURL(fullSize), http.StatusFound)
}

// ServeHeartbeatCheck serves the heartbeat check-in endpoint.
func (h *Handler) ServeHeartbeatCheck(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	monitorID := parts[len(parts)-1]

	err := retry.DoTemporaryError(func(_ int) error {
		return h.c.HeartbeatStore.RecordHeartbeat(ctx, monitorID)
	},
		retry.Log(ctx),
		retry.Limit(12),
		retry.FibBackoff(time.Second),
	)
	if errors.Is(err, sql.ErrNoRows) {
		auth.Delay(ctx)
		http.NotFound(w, r)
		return
	}
	if errutil.HTTPError(ctx, w, err) {
		return
	}
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

	meta := make(map[string]string)
	for _, v := range r.Form["meta"] {
		key, val, ok := strings.Cut(v, "=")
		if !ok {
			continue
		}
		meta[key] = val
	}

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct == "application/json" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var b struct {
			Summary, Details, Action, Dedup *string
			Meta                            map[string]string
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
		if b.Meta != nil {
			meta = b.Meta
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
		Source:    alert.SourceGeneric,
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
		createdAlert, isNew, err := h.c.AlertStore.CreateOrUpdateWithMeta(ctx, a, meta)
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
