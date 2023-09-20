package notifyapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
)

// Handler responds to generic API requests
type Handler struct {
	c Config
}

// NewHandler creates a new Handler, registering generic API routes using chi.
func NewHandler(c Config) *Handler {
	return &Handler{c: c}
}

// func hash(b []byte) uint64 {
// 	h := fnv.New64a()
// 	h.Write(b)
// 	return h.Sum64()
// }

// ServeCreateAlert allows creating or closing an alert.
func (h *Handler) ServeCreateAlert(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := permission.LimitCheckAny(ctx, permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	// serviceID := permission.ServiceID(ctx)

	// action := r.FormValue("action")

	b := make(map[string]interface{})
	// var dedupeHash uint64

	ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if ct == "application/json" {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = json.Unmarshal(data, &b)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// if b["action"] != nil {
		// 	actionValue, ok := b["action"].(string)
		// 	if !ok {
		// 		http.Error(w, fmt.Sprintf("Field 'action' is not a string: %s", b["action"]), http.StatusBadRequest)
		// 		return
		// 	}
		// 	action = actionValue
		// 	delete(b, "action")
		// 	data, err = json.Marshal(&b)
		// }
		// dedupeHash = hash(data)
	}

	integrationKey := r.URL.Query().Get("token")
	rules, err := h.FindMatchingRules(ctx, integrationKey, b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(rules) == 0 {
		http.Error(w, fmt.Sprintf("No rules found for integration key: %s", integrationKey), http.StatusAccepted)
		return
	}

	// var summary string
	// var details string
	// for _, rule := range rules {
	// 	for _, action := range rule.Actions {
	// 		summary += fmt.Sprintf("%s", action.Destination)
	// 		details += action.Message
	// 	}
	// }

	// status := alert.StatusTriggered
	// if action == "close" {
	// 	status = alert.StatusClosed
	// }

	// summary = validate.SanitizeText(summary, alert.MaxSummaryLength)
	// details = validate.SanitizeText(details, alert.MaxDetailsLength)

	// a := &alert.Alert{
	// 	Summary:   summary,
	// 	Details:   details,
	// 	Source:    alert.SourceNotify,
	// 	ServiceID: serviceID,
	// 	Dedup:     alert.NewUserDedup(fmt.Sprintf("%d", dedupeHash)),
	// 	Status:    status,
	// }

	// var resp struct {
	// 	AlertID   int
	// 	ServiceID string
	// 	IsNew     bool
	// }

	// err = retry.DoTemporaryError(func(int) error {
	// 	createdAlert, isNew, err := h.c.AlertStore.CreateOrUpdate(ctx, a)
	// 	if createdAlert != nil {
	// 		resp.AlertID = createdAlert.ID
	// 		resp.ServiceID = createdAlert.ServiceID
	// 		resp.IsNew = isNew
	// 	}

	// 	return err
	// },
	// 	retry.Log(ctx),
	// 	retry.Limit(10),
	// 	retry.FibBackoff(time.Second),
	// )
	// if errutil.HTTPError(ctx, w, errors.Wrap(err, "create alert")) {
	// 	return
	// }

	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(204)
		return
	}

	data, err := json.Marshal(rules)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}
