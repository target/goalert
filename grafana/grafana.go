package grafana

import (
	"encoding/json"
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

type grafanaPost struct {
	RuleName string
	RuleID   int
	Message  string
	State    string
	Title    string
	RuleURL  string
}

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

func GrafanaToEventsAPI(aDB alert.Store, intDB integrationkey.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		err := permission.LimitCheckAny(ctx, permission.Service)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		serviceID := permission.ServiceID(ctx)

		var g grafanaPost
		err = json.NewDecoder(r.Body).Decode(&g)
		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from grafana: %v", err)
			return
		}

		ctx = log.WithFields(ctx, log.Fields{
			"RuleURL": g.RuleURL,
			"State":   g.State,
		})

		var grafanaState alert.Status
		switch g.State {
		case "alerting":
			grafanaState = alert.StatusTriggered
		case "ok":
			grafanaState = alert.StatusClosed
		case "no_data":
			// no data..
			return
		default:
			log.Logf(ctx, "bad request from grafana: missing or invalid state")
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		var urlStr string
		if validate.AbsoluteURL("RuleURL", g.RuleURL) == nil {
			urlStr = g.RuleURL
		}
		body := strings.TrimSpace(urlStr + "\n\n" + g.Message)

		//dedupe is description, source, and serviceID
		msg := &alert.Alert{
			Summary:   validate.SanitizeText(g.RuleName, alert.MaxSummaryLength),
			Details:   validate.SanitizeText(body, alert.MaxDetailsLength),
			Status:    grafanaState,
			Source:    alert.SourceGrafana,
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
		if errutil.HTTPError(ctx, w, errors.Wrap(err, "create or update alert for grafana")) {
			return
		}
	}
}
