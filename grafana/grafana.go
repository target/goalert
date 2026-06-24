package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
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

var detailsTmpl = template.Must(template.New("details").Funcs(template.FuncMap{
	"escapeTableCell": func(s string) string {
		s = strings.ReplaceAll(s, "\n", "<br />")
		s = strings.ReplaceAll(s, "|", "\\|")
		return s
	},
	"codeBlock": func(s string) string {
		delim := "```"
		for strings.Contains(s, delim) {
			delim += "`"
		}

		return delim + "\n" + s + "\n" + delim
	},
}).Parse(`
{{- if .Labels }}
| Label | Value |
| ----- | ----- |
{{- range $k, $v := .Labels }}
| {{ $k }} | {{escapeTableCell $v }} |
{{- end }}
{{- end }}


{{- if .Annotations }}
| Annotation | Value |
| ---------- | ----- |
{{- range $k, $v := .Annotations }}
| {{ $k }} | {{escapeTableCell $v }} |
{{- end }}
{{- end }}


{{if .GeneratorURL}}Source: {{ .GeneratorURL }}{{end}}

{{if .SlienceURL}}Silence: {{ .SlienceURL }}{{end}}

{{if .ImageURL}}![Panel Snapshot]({{ .ImageURL }}){{end}}

{{codeBlock .ValueString }}
`))

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

func alertsFromLegacy(ctx context.Context, req *http.Request, serviceID string, data []byte) ([]alert.Alert, error) {
	var g struct {
		RuleName string
		RuleID   int
		Message  string
		State    string
		Title    string
		RuleURL  string
		ImageURL string
	}
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}

	var grafanaState alert.Status
	switch g.State {
	case "alerting":
		grafanaState = alert.StatusTriggered
	case "ok":
		grafanaState = alert.StatusClosed
	case "no_data":
		// no data..
		return nil, nil
	default:
		return nil, errors.Errorf("grafana: unknown state: %s", g.State)
	}

	var urlStr string
	if validate.AbsoluteURL("RuleURL", g.RuleURL) == nil {
		urlStr = g.RuleURL
	}
	body := strings.TrimSpace(urlStr + "\n\n" + g.Message)

	if validate.AbsoluteURL("ImageURL", g.ImageURL) == nil {
		body += "\n\n![Panel Snapshot](" + g.ImageURL + ")"
	}

	// dedupe is description, source, and serviceID
	return []alert.Alert{{
		Summary:   validate.SanitizeText(g.RuleName, alert.MaxSummaryLength),
		Details:   validate.SanitizeText(body, alert.MaxDetailsLength),
		Status:    grafanaState,
		ServiceID: serviceID,
		Source:    alert.SourceGrafana,
		Dedup:     alert.NewUserDedup(req.FormValue("dedup")),
	}}, nil
}

func alertsFromV1(ctx context.Context, serviceID string, data []byte) ([]alert.Alert, error) {
	var g struct {
		Alerts []struct {
			Status              string
			Labels, Annotations map[string]string
			ValueString         string
			Fingerprint         string
			GeneratorURL        string
			SlienceURL          string
			ImageURL            string
		}
	}
	err := json.Unmarshal(data, &g)
	if err != nil {
		return nil, err
	}

	var alerts []alert.Alert
	for _, a := range g.Alerts {
		var alertStatus alert.Status
		switch a.Status {
		case "firing":
			alertStatus = alert.StatusTriggered
		case "resolved":
			alertStatus = alert.StatusClosed
		default:
			return nil, errors.Errorf("grafana: unknown status: %s", a.Status)
		}

		var buf strings.Builder
		err := detailsTmpl.Execute(&buf, a)
		if err != nil {
			return nil, err
		}
		summary := a.Annotations["summary"]
		if summary == "" {
			summary = a.Labels["alertname"]
		}

		alerts = append(alerts, alert.Alert{
			Summary:   validate.SanitizeText(summary, alert.MaxSummaryLength),
			Details:   validate.SanitizeText(buf.String(), alert.MaxDetailsLength),
			Status:    alertStatus,
			ServiceID: serviceID,
			Source:    alert.SourceGrafana,
			Dedup:     alert.NewUserDedup(a.Fingerprint),
		})
	}

	return alerts, nil
}

func GrafanaToEventsAPI(aDB *alert.Store, intDB *integrationkey.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		err := permission.LimitCheckAny(ctx, permission.Service)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		serviceID := permission.ServiceID(ctx)

		data, err := io.ReadAll(r.Body)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		var versionInfo struct{ Version string }
		err = json.Unmarshal(data, &versionInfo)
		if clientError(w, http.StatusBadRequest, err) {
			return
		}

		var alerts []alert.Alert
		switch versionInfo.Version {
		case "1":
			alerts, err = alertsFromV1(ctx, serviceID, data)
		case "":
			alerts, err = alertsFromLegacy(ctx, r, serviceID, data)
		default:
			clientError(w, http.StatusBadRequest, errors.Errorf("grafana: unknown payload version: %s", versionInfo.Version))
			return
		}

		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from grafana: %v", err)
			return
		}
		if len(alerts) == 0 {
			// no data
			return
		}
		if len(alerts) > 10 {
			log.Log(ctx, fmt.Errorf("grafana: too many alerts (truncating to 10): %d", len(alerts)))
			alerts = alerts[:10]
		}

		var hasFailures bool
		for _, a := range alerts {
			err = retry.DoTemporaryError(func(int) error {
				_, _, err = aDB.CreateOrUpdate(ctx, &a)
				return err
			},
				retry.Log(ctx),
				retry.Limit(10),
				retry.FibBackoff(time.Second),
			)
			if err != nil {
				log.Log(ctx, fmt.Errorf("grafana: create alert: %w", err))
				hasFailures = true
			}
		}

		if hasFailures {
			http.Error(w, "failed to create alerts", http.StatusInternalServerError)
			return
		}
	}
}
