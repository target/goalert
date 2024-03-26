package prometheus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

/* Example payload

```
{
  "receiver": "goalert",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "InstanceDown",
        "code": "200",
        "instance": "127.0.0.1:9090",
        "job": "prometheus",
        "monitor": "codelab-monitor",
        "severity": "critical"
      },
      "annotations": {
        "details": "127.0.0.1:9090 of job prometheus has been down for more than 1 minute.",
        "summary": "Instance 127.0.0.1:9090 down"
      },
      "startsAt": "2020-08-08T14:32:08.326990857Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://pop-os:9090/graph?g0.expr=promhttp_metric_handler_requests_total+%3E+20\u0026g0.tab=1",
      "fingerprint": "791cec13fcba0368"
    },
    {
      "status": "firing",
      "labels": {
        "alertname": "InstanceDown",
        "code": "200",
        "instance": "localhost:9090",
        "job": "prometheus",
        "monitor": "codelab-monitor",
        "severity": "critical"
      },
      "annotations": {
        "details": "localhost:9090 of job prometheus has been down for more than 1 minute.",
        "summary": "Instance localhost:9090 down"
      },
      "startsAt": "2020-08-08T02:21:08.326990857Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://pop-os:9090/graph?g0.expr=promhttp_metric_handler_requests_total+%3E+20\u0026g0.tab=1",
      "fingerprint": "8df98227bdd81384"
    }
  ],
  "groupLabels": {},
  "commonLabels": {
    "alertname": "InstanceDown",
    "code": "200",
    "job": "prometheus",
    "monitor": "codelab-monitor",
    "severity": "critical"
  },
  "commonAnnotations": {},
  "externalURL": "http://pop-os:9093",
  "version": "4",
  "groupKey": "{}:{}",
  "truncatedAlerts": 0
}
```
*/

type postBody struct {
	Status      string
	ExternalURL string

	Alerts []postBodyAlert

	CommonLabels struct {
		Instance  string
		AlertName string `json:"alertname"`
	}

	CommonAnnotations struct {
		Summary string
		Details string
	}
}
type postBodyAlert struct {
	Labels struct {
		AlertName string
		Instance  string
	}
	Annotations struct {
		Summary string
		Details string
	}
	GeneratorURL string
}

func (a postBodyAlert) Summary() string {
	if a.Annotations.Summary != "" {
		return a.Annotations.Summary
	}

	return a.Labels.AlertName + " " + a.Labels.Instance
}
func (a postBodyAlert) gen() string {
	if a.GeneratorURL == "" {
		return ""
	}

	return fmt.Sprintf(" [View](%s)", a.GeneratorURL)
}
func (a postBodyAlert) Details() string {
	if a.Annotations.Details != "" {
		return a.Annotations.Details + a.gen()
	}

	return a.Summary() + a.gen()
}
func (b postBody) Summary() string {
	if b.CommonAnnotations.Summary != "" {
		return b.CommonAnnotations.Summary
	}
	if b.CommonLabels.AlertName == "" {
		// different alerts
		return b.Alerts[0].Summary() + fmt.Sprintf(" and %d others", len(b.Alerts)-1)
	}

	// we have a common alert name
	if b.CommonLabels.Instance != "" {
		return b.CommonLabels.AlertName + " " + b.CommonLabels.Instance
	}

	var instances []string
	for _, a := range b.Alerts {
		instances = append(instances, a.Labels.Instance)
	}

	return b.CommonLabels.AlertName + " " + strings.Join(instances, ",")
}

func (b postBody) Details(payload string) string {
	var s strings.Builder
	if b.ExternalURL != "" {
		fmt.Fprintf(&s, "[Prometheus Alertmanager UI](%s)\n\n", b.ExternalURL)
	}
	if b.CommonAnnotations.Details != "" {
		s.WriteString(b.CommonAnnotations.Details + "\n\n")
	} else {
		for _, a := range b.Alerts {
			s.WriteString(a.Details() + "\n\n")
		}
	}
	if payload != "" {
		fmt.Fprintf(&s, "## Payload\n\n```json\n%s\n```\n", payload)
	}
	return s.String()
}

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

func PrometheusAlertmanagerEventsAPI(aDB *alert.Store, intDB *integrationkey.Store) http.HandlerFunc {

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

		data := make([]byte, buf.Len())
		copy(data, buf.Bytes())
		buf.Reset()
		err = json.Indent(&buf, data, "", "  ")
		if err == nil {
			data = buf.Bytes()
		}

		summary := validate.SanitizeText(body.Summary(), alert.MaxSummaryLength)
		msg := &alert.Alert{
			Summary:   summary,
			Details:   validate.SanitizeText(body.Details(string(data)), alert.MaxDetailsLength),
			Status:    status,
			Source:    alert.SourcePrometheusAlertmanager,
			ServiceID: serviceID,
			Dedup:     alert.NewUserDedup(summary),
		}

		err = retry.DoTemporaryError(func(int) error {
			_, _, err = aDB.CreateOrUpdate(ctx, msg, nil)
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
