package prometheus

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

type alertmanagerPost struct {
	GroupKey          string                   `json:"groupKey"`
	Status            string                   `json:"status"`
	Receiver          string                   `json:"receiver"`
	GroupLabels       map[string]interface{}   `json:"groupLabels"`
	CommonLabels      map[string]interface{}   `json:"commonLabels"`
	CommonAnnotations map[string]interface{}   `json:"commonAnnotations"`
	ExternalURL       string                   `json:"externalURL"`
	Alerts            []map[string]interface{} `json:"alerts"`
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

		var alertmanager alertmanagerPost
		err = json.NewDecoder(r.Body).Decode(&alertmanager)
		if clientError(w, http.StatusBadRequest, err) {
			log.Logf(ctx, "bad request from prometheus alertmanager: %v", err)
			return
		}

		ctx = log.WithFields(ctx, log.Fields{
			"RuleURL": alertmanager.ExternalURL,
			"Status":  alertmanager.Status,
		})

		var alertmanagerState alert.Status
		switch alertmanager.Status {
		case "FIRING":
			alertmanagerState = alert.StatusTriggered
		case "RESOLVED":
			alertmanagerState = alert.StatusClosed
		default:
			log.Logf(ctx, "bad request from prometheus alertmanager: missing or invalid state")
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		var urlStr string
		if validate.AbsoluteURL("externalURL", alertmanager.ExternalURL) == nil {
			urlStr = alertmanager.ExternalURL
		}
		body := strings.TrimSpace(urlStr + "\n\n" + alertmanager.Receiver)

		//dedupe is description, source, and serviceID
		msg := &alert.Alert{
			Summary:   validate.SanitizeText(alertmanager.Receiver, alert.MaxSummaryLength),
			Details:   validate.SanitizeText(body, alert.MaxDetailsLength),
			Status:    alertmanagerState,
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
