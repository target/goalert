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

	info, err := gadb.NewCompat(s.db).CalSubRenderInfo(ctx, uuid.MustParse(src.ID))
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	shifts, err := s.oc.HistoryBySchedule(ctx, info.ScheduleID.String(), info.Now, info.Now.AddDate(1, 0, 0))
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	var subCfg SubscriptionConfig
	err = json.Unmarshal(info.Config, &subCfg)
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	if !subCfg.FullSchedule {
		// filter out other users
		filtered := shifts[:0]
		for _, s := range shifts {
			if s.UserID != info.UserID.String() {
				continue
			}
			filtered = append(filtered, s)
		}
		shifts = filtered
	}

	data := renderData{
		ApplicationName: cfg.ApplicationName(),
		ScheduleID:      info.ScheduleID,
		ScheduleName:    info.ScheduleName,
		Shifts:          shifts,
		ReminderMinutes: subCfg.ReminderMinutes,
		Version:         version.GitVersion(),
		GeneratedAt:     info.Now,
		FullSchedule:    subCfg.FullSchedule,
	}

	if subCfg.FullSchedule {
		// When rendering the full schedule, we need to fetch the names of all users.
		data.UserNames = make(map[string]string)
		var uniqueIDs []uuid.UUID
		for _, s := range shifts {

			// We'll use the map to track which IDs we've already seen.
			// That way we don't ask the DB for the same user multiple times.
			if _, ok := data.UserNames[s.UserID]; ok {
				continue
			}
			data.UserNames[s.UserID] = "Unknown User"
			uniqueIDs = append(uniqueIDs, uuid.MustParse(s.UserID))
		}

		users, err := gadb.NewCompat(s.db).CalSubUserNames(ctx, uniqueIDs)
		if errutil.HTTPError(ctx, w, err) {
			return
		}

		for _, u := range users {
			data.UserNames[u.ID.String()] = u.Name
		}
	}

	calData, err := data.renderICal()
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	w.Header().Set("Content-Type", "text/calendar")
	_, _ = w.Write(calData)
}
