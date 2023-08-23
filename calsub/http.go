package calsub

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/version"
)

// ServeICalData will return an iCal file for the subscription associated with the current request.
func (s *Store) ServeICalData(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	src := permission.Source(ctx)
	cfg := config.FromContext(ctx)
	if src.Type != permission.SourceTypeCalendarSubscription || cfg.General.DisableCalendarSubscriptions {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	info, err := gadb.New(s.db).CalSubRenderInfo(ctx, uuid.MustParse(src.ID))
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	var userIDs []string
	userIDs = append(userIDs, info.UserID.String())
	shifts, err := s.oc.HistoryBySchedule(ctx, info.ScheduleID.String(), info.Now, info.Now.AddDate(1, 0, 0), userIDs)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	var subCfg SubscriptionConfig
	err = json.Unmarshal(info.Config, &subCfg)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	data := renderData{
		ApplicationName: cfg.ApplicationName(),
		ScheduleID:      info.ScheduleID,
		ScheduleName:    info.ScheduleName,
		Shifts:          shifts,
		ReminderMinutes: subCfg.ReminderMinutes,
		Version:         version.GitVersion(),
		GeneratedAt:     info.Now,
	}

	calData, err := data.renderICal()
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	w.Header().Set("Content-Type", "text/calendar")
	_, _ = w.Write(calData)
}
