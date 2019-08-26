package genericapi

import (
	"database/sql"
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
	if errors.Cause(err) == sql.ErrNoRows {
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
		return h.c.HeartbeatStore.Heartbeat(ctx, monitorID)
	},
		retry.Log(ctx),
		retry.Limit(12),
		retry.FibBackoff(time.Second),
	)
	if errors.Cause(err) == sql.ErrNoRows {
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

	summary = validate.SanitizeText(summary, alert.MaxSummaryLength)
	details = validate.SanitizeText(details, alert.MaxDetailsLength)

	status := alert.StatusTriggered
	if r.FormValue("action") == "close" {
		status = alert.StatusClosed
	}

	a := &alert.Alert{
		Summary:   summary,
		Details:   details,
		Source:    alert.SourceGeneric,
		ServiceID: serviceID,
		Dedup:     alert.NewUserDedup(r.FormValue("dedup")),
		Status:    status,
	}

	err = retry.DoTemporaryError(func(int) error {
		_, err = h.c.AlertStore.CreateOrUpdate(ctx, a)
		return err
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second),
	)
	if errutil.HTTPError(ctx, w, errors.Wrap(err, "create alert")) {
		return
	}

	w.WriteHeader(204)
}
