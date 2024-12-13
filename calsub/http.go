package calsub

import (
	"context"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/oncall"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/version"
)

// PayloadType is the embedded type & version for calendar subscription payloads.
const PayloadType = "calendar-subscription/v1"

// JSONResponseV1 is the JSON response format for calendar subscription requests.
type JSONResponseV1 struct {
	AppName    string
	AppVersion string

	// Type is the embedded type & version for calendar subscription payloads and should be set to PayloadType.
	Type string

	ScheduleID   uuid.UUID
	ScheduleName string
	ScheduleURL  string

	Start, End time.Time

	Shifts []JSONShiftV1
}

// JSONShiftV1 is the JSON response format for a shift in a calendar subscription.
type JSONShiftV1 struct {
	Start, End time.Time

	UserID   uuid.UUID
	UserName string
	UserURL  string

	Truncated bool
}

func (s *Store) userNameMap(ctx context.Context, shifts []oncall.Shift) (map[string]string, error) {
	names := make(map[string]string)
	var uniqueIDs []uuid.UUID
	for _, s := range shifts {

		// We'll use the map to track which IDs we've already seen.
		// That way we don't ask the DB for the same user multiple times.
		if _, ok := names[s.UserID]; ok {
			continue
		}
		names[s.UserID] = "Unknown User"
		uniqueIDs = append(uniqueIDs, uuid.MustParse(s.UserID))
	}

	users, err := gadb.New(s.db).CalSubUserNames(ctx, uniqueIDs)
	if err != nil {
		return nil, fmt.Errorf("lookup user names: %w", err)
	}

	for _, u := range users {
		names[u.ID.String()] = u.Name
	}
	return names, nil
}

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

	ct, _, _ := mime.ParseMediaType(req.Header.Get("Accept"))
	if ct == "application/json" {
		data := JSONResponseV1{
			AppName:      cfg.ApplicationName(),
			AppVersion:   version.GitVersion(),
			Type:         PayloadType,
			ScheduleID:   info.ScheduleID,
			ScheduleName: info.ScheduleName,
			ScheduleURL:  cfg.CallbackURL("/schedules/" + info.ScheduleID.String()),
			Start:        info.Now,
			End:          info.Now.AddDate(1, 0, 0),
		}
		m, err := s.userNameMap(ctx, shifts)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
		for _, s := range shifts {
			data.Shifts = append(data.Shifts, JSONShiftV1{
				Start:     s.Start,
				End:       s.End,
				Truncated: s.Truncated,
				UserID:    uuid.MustParse(s.UserID),
				UserName:  m[s.UserID],
				UserURL:   cfg.CallbackURL("/users/" + s.UserID),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(data)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
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
		FullSchedule:    subCfg.FullSchedule,
	}

	if subCfg.FullSchedule {
		// When rendering the full schedule, we need to fetch the names of all users.
		data.UserNames, err = s.userNameMap(ctx, shifts)
		if errutil.HTTPError(ctx, w, err) {
			return
		}
	}

	calData, err := data.renderICal()
	if errutil.HTTPError(ctx, w, err) {
		return
	}

	w.Header().Set("Content-Type", "text/calendar")
	_, _ = w.Write(calData)
}
