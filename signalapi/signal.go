package signalapi

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"text/template"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
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
type SignalAlertMapper struct {
	Summary, Details, Action, Dedup string
	Status                          alert.Status
}

// ContactMethod types
const (
	DestTypeAlert = "ALERT"
)

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
		for _, action := range rule.Actions {
			if rule.SendAlert && strings.EqualFold(action.DestType, DestTypeAlert) {
				sigAlert, err := buildOutgoingAlertPayload(action, requestBody)
				if errutil.HTTPError(ctx, w, errors.Wrap(err, "create alert")) {
					return
				}

				a := &alert.Alert{
					Summary:   sigAlert.Summary,
					Details:   sigAlert.Details,
					Source:    alert.SourceSignal,
					ServiceID: serviceID,
					Dedup:     alert.NewUserDedup(sigAlert.Dedup),
					Status:    sigAlert.Status,
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
				continue
			}
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

func buildOutgoingAlertPayload(action rule.Action, incomingPayload map[string]interface{}) (sigAlert SignalAlertMapper, err error) {
	if incomingPayload["action"] != nil {
		val, ok := incomingPayload["action"].(string)
		if !ok {
			return sigAlert, fmt.Errorf("Field 'action' is not a string: %s", incomingPayload["action"])
		}
		sigAlert.Action = val
	}
	if incomingPayload["summary"] != nil {
		val, ok := incomingPayload["summary"].(string)
		if !ok {
			return sigAlert, fmt.Errorf("Field 'summary' is not a string: %s", incomingPayload["summary"])
		}
		sigAlert.Summary = val
	} else {
		var val string
		for _, content := range action.Contents {
			if content.Prop == "summary" {
				val, err = InjectTemplateValues(content.Value, incomingPayload)
				if err != nil {
					return sigAlert, fmt.Errorf("Error creating 'summary' from template: %s", err)
				}
			}
		}
		sigAlert.Summary = val
	}
	if incomingPayload["details"] != nil {
		val, ok := incomingPayload["details"].(string)
		if !ok {
			return sigAlert, fmt.Errorf("Field 'details' is not a string: %s", incomingPayload["details"])
		}
		sigAlert.Details = val
	} else {
		var val string
		for _, content := range action.Contents {
			if content.Prop == "details" {
				val, err = InjectTemplateValues(content.Value, incomingPayload)
				if err != nil {
					return sigAlert, fmt.Errorf("Error creating 'details' from template: %s", err)
				}
			}
		}
		sigAlert.Summary = val
	}
	if incomingPayload["dedup"] != nil {
		val, ok := incomingPayload["dedup"].(string)
		if !ok {
			return sigAlert, fmt.Errorf("Field 'dedup' is not a string: %s", incomingPayload["dedup"])
		}
		sigAlert.Dedup = val
	}

	sigAlert.Status = alert.StatusTriggered
	if sigAlert.Action == "close" {
		sigAlert.Status = alert.StatusClosed
	}

	return sigAlert, nil
}

func InjectTemplateValues(tmplStr string, data map[string]interface{}) (string, error) {
	tmpl, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return "", err
	}
	return result.String(), nil
}
