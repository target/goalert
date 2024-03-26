package site24x7

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

type post struct {
	MonitorDashboardURL string `json:"MONITOR_DASHBOARD_LINK"` // using URL instead of Link to match fields used in GoAlert, we can just map it to the JSON name
	Status              string `json:"STATUS"`
	MonitorName         string `json:"MONITORNAME"`
}

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

func Site24x7ToEventsAPI(aDB *alert.Store, intDB *integrationkey.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		err := permission.LimitCheckAny(ctx, permission.Service)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		serviceID := permission.ServiceID(ctx)

		var g post
		err = json.NewDecoder(r.Body).Decode(&g)
		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from site24x7: %v", err)
			return
		}

		ctx = log.WithFields(ctx, log.Fields{
			"RuleURL": g.MonitorDashboardURL,
			"State":   g.Status,
		})

		var site24x7State alert.Status
		switch g.Status {
		case "DOWN", "CRITICAL", "TROUBLE":
			site24x7State = alert.StatusTriggered
		case "UP":
			site24x7State = alert.StatusClosed
		default:
			log.Logf(ctx, "bad request from site24x7: missing or invalid state")
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		var urlStr string
		if validate.AbsoluteURL("MONITOR_DASHBOARD_LINK", g.MonitorDashboardURL) == nil {
			urlStr = g.MonitorDashboardURL
		}
		body := strings.TrimSpace(urlStr + "\n\n" + g.MonitorName)

		//dedupe is description, source, and serviceID
		msg := &alert.Alert{
			Summary:   validate.SanitizeText(g.MonitorName, alert.MaxSummaryLength),
			Details:   validate.SanitizeText(body, alert.MaxDetailsLength),
			Status:    site24x7State,
			Source:    alert.SourceSite24x7,
			ServiceID: serviceID,
			Dedup:     alert.NewUserDedup(r.FormValue("dedup")),
		}

		err = retry.DoTemporaryError(func(int) error {
			_, _, err = aDB.CreateOrUpdate(ctx, msg, nil)
			return err
		},
			retry.Log(ctx),
			retry.Limit(10),
			retry.FibBackoff(time.Second),
		)
		if errutil.HTTPError(ctx, w, errors.Wrap(err, "create or update alert for site24x7")) {
			return
		}
	}
}
