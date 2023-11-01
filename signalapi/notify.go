package signalapi

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/service/rule"
	"github.com/target/goalert/signal"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/sqlutil"
)

// Handler responds to generic API requests
type Handler struct {
	c Config
}

// NewHandler creates a new Handler, registering generic API routes using chi.
func NewHandler(c Config) *Handler {
	return &Handler{c: c}
}

// ServeCreateSignals allows creating signals.
func (h *Handler) ServeCreateSignals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := permission.LimitCheckAny(ctx, permission.Service)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

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

	integrationKey := r.URL.Query().Get("token")
	rules, err := h.findMatchingRules(ctx, integrationKey, requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(rules) == 0 {
		if r.Header.Get("Accept") != "application/json" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("[]"))
		return
	}

	signals := []signal.Signal{}

	for _, rule := range rules {
		if rule.SendAlert {
			// TODO: implement create alert logic
			continue
		}
		for _, action := range rule.Actions {
			signals = append(signals, signal.Signal{
				ServiceID:       serviceID,
				ServiceRuleID:   rule.ID,
				OutgoingPayload: buildOutgoingPayload(action, requestBody),
			})
		}
	}

	createdSignals := []*signal.Signal{}

	tx, err := h.c.DB.BeginTx(ctx, nil)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	defer sqlutil.Rollback(ctx, "signal: create many", tx)
	err = retry.DoTemporaryError(func(int) error {
		createdSignals, err = h.c.SignalStore.CreateMany(ctx, tx, signals)
		return err
	},
		retry.Log(ctx),
		retry.Limit(10),
		retry.FibBackoff(time.Second),
	)
	if errutil.HTTPError(ctx, w, errors.Wrap(err, "create signals")) {
		return
	}

	err = tx.Commit()
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	if r.Header.Get("Accept") != "application/json" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(createdSignals)
	if errutil.HTTPError(ctx, w, err) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// buildOutgoingPayload is a hypothetical implementation.
// TODO: update when plugin implementation has been defined
func buildOutgoingPayload(action rule.Action, incomingPayload map[string]interface{}) map[string]interface{} {
	outgoingPayload := make(map[string]interface{})

	if action.DestType != "" {
		outgoingPayload["dest_type"] = action.DestType
	}

	outgoingPayload["received_payload"] = incomingPayload

	return outgoingPayload
}
