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

type site24x7Post struct {
	MONITOR_DASHBOARD_LINK  string
	MONITORTYPE             string
	STATUS                  string
	REASON                  string
	MONITORNAME             string
	FAILED_LOCATIONS        string
	INCIDENT_REASON         string
	OUTAGE_TIME_UNIX_FORMAT string
	MONITORURL              string
	MONITOR_GROUPNAME       string
	INCIDENT_TIME           string
	INCIDENT_TIME_ISO       string
	RCA_LINK                string
	//ct                      int
}

func clientError(w http.ResponseWriter, code int, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, http.StatusText(code), code)
	return true
}

func Site24x7ToEventsAPI(aDB alert.Store, intDB integrationkey.Store) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		err := permission.LimitCheckAny(ctx, permission.Service)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		serviceID := permission.ServiceID(ctx)

		var g site24x7Post
		err = json.NewDecoder(r.Body).Decode(&g)
		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from site24x7: %v", err)
			return
		}

		ctx = log.WithFields(ctx, log.Fields{
			"RuleURL": g.MONITOR_DASHBOARD_LINK,
			"State":   g.STATUS,
		})

		var site24x7State alert.Status
		switch g.STATUS {
		case "DOWN":
			site24x7State = alert.StatusTriggered
		case "CRITICAL":
			site24x7State = alert.StatusTriggered
		case "TROUBLE":
			site24x7State = alert.StatusTriggered
		case "UP":
			site24x7State = alert.StatusClosed
		default:
			log.Logf(ctx, "bad request from site24x7: missing or invalid state")
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		var urlStr string
		if validate.AbsoluteURL("RuleURL", g.MONITOR_DASHBOARD_LINK) == nil {
			urlStr = g.MONITOR_DASHBOARD_LINK
		}
		body := strings.TrimSpace(urlStr + "\n\n" + g.MONITORNAME)

		//dedupe is description, source, and serviceID
		msg := &alert.Alert{
			Summary:   validate.SanitizeText(g.MONITORNAME, alert.MaxSummaryLength),
			Details:   validate.SanitizeText(body, alert.MaxDetailsLength),
			Status:    site24x7State,
			Source:    alert.SourceSite24x7,
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
		if errutil.HTTPError(ctx, w, errors.Wrap(err, "create or update alert for site24x7")) {
			return
		}
	}
}
