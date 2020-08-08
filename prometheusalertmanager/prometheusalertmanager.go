package prometheus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/integrationkey"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"
)

type postBody struct {
	Status      string
	ExternalURL string

	CommonLabels struct {
		Instance  string
		AlertName string `json:"alertname"`
	}

	CommonAnnotations struct {
		Summary string
		Details string
	}
}

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

func PrometheusAlertmanagerEventsAPI(aDB alert.Store, intDB integrationkey.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		err := permission.LimitCheckAny(ctx, permission.Service)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		serviceID := permission.ServiceID(ctx)

		var body postBody
		var buf bytes.Buffer
		err = json.NewDecoder(io.TeeReader(r.Body, &buf)).Decode(&body)
		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from prometheus alertmanager: %v", err)
			return
		}

		ctx = log.WithFields(ctx, log.Fields{
			"RuleURL": body.ExternalURL,
			"Status":  body.Status,
		})

		var status alert.Status
		switch body.Status {
		case "firing":
			status = alert.StatusTriggered
		case "resolved":
			status = alert.StatusClosed
		default:
			log.Logf(ctx, "bad request from prometheus alertmanager: missing or invalid state")
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		summary := body.CommonAnnotations.Summary
		if summary == "" {
			summary = fmt.Sprintf("Alertmanager: %s %s", body.CommonLabels.AlertName, body.CommonLabels.Instance)
		}

		data := make([]byte, buf.Len())
		copy(data, buf.Bytes())
		buf.Reset()
		err = json.Indent(&buf, data, "", "  ")
		if err == nil {
			data = buf.Bytes()
		}

		details := fmt.Sprintf("%s\n\n[Alertmanager](%s)\n\n## Payload\n\n```\n%s\n```",
			body.CommonAnnotations.Details,
			body.ExternalURL,
			string(data),
		)

		msg := &alert.Alert{
			Summary:   validate.SanitizeText(summary, alert.MaxSummaryLength),
			Details:   validate.SanitizeText(details, alert.MaxDetailsLength),
			Status:    status,
			Source:    alert.SourcePrometheusAlertmanager,
			ServiceID: serviceID,
			Dedup:     alert.NewUserDedup(r.FormValue("dedup")),
		}

		err = retry.DoTemporaryError(func(int) error {
			_, err = aDB.CreateOrUpdate(ctx, msg)
			return err
		},
			retry.Log(ctx),
			retry.Limit(10),
			retry.FibBackoff(time.Second),
		)
		if errutil.HTTPError(ctx, w, errors.Wrap(err, "create or update alert for prometheus alertmanager")) {
			return
		}
	}
}
