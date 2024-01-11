package universalapi

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

// Handler responds to universal API requests
type Handler struct {
	c Config
}

// NewHandler creates a new Handler, registering universal API routes using chi.
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

	intkey := permission.Source(ctx).ID

	serviceID := permission.ServiceID(ctx)
	requestBody := make(map[string]interface{})

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct == "application/json" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(data, &requestBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	// gets the ruleset of the corresponding intkey stopping after finding the first rule
	rules, err := h.FetchRules(ctx, intkey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// gets the list of rules which matches with the request body
	matchedRules := MatchRules(ctx, rules, requestBody, nil)

	if len(matchedRules) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for _, rule := range matchedRules {
		outgoingAlert, err := BuildOutgoingAlert(requestBody, rule)
		if errutil.HTTPError(ctx, w, errors.Wrap(err, "create alert")) {
			return
		}

		summary := validate.SanitizeText(outgoingAlert.Summary, alert.MaxSummaryLength)
		details := validate.SanitizeText(outgoingAlert.Details, alert.MaxDetailsLength)

		a := &alert.Alert{
			Summary:   summary,
			Details:   details,
			Source:    alert.SourceUniversal,
			ServiceID: serviceID,
			Dedup:     alert.NewUserDedup(rule.Dedup),
			Status:    outgoingAlert.Status,
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
	}

	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
