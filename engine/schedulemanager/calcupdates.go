package schedulemanager

import (
	"encoding/json"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/jsonutil"
)

// overrideSpan is a user override as the schedule manager sees it: who is
// added and/or removed, and the window during which it applies.
//
// Both the batch (SchedMgrOverrides) and per-schedule
// (SchedMgrOverridesForSchedule) row types convert to this one so the on-call
// calculation does not depend on which query produced the rows.
type overrideSpan struct {
	AddUserID, RemoveUserID uuid.NullUUID
	Start, End              time.Time
}

// overrideSpanFromRow converts a batch override row into an overrideSpan.
func overrideSpanFromRow(row gadb.SchedMgrOverridesRow) overrideSpan {
	return overrideSpan{
		AddUserID:    row.AddUserID,
		RemoveUserID: row.RemoveUserID,
		Start:        row.StartTime,
		End:          row.EndTime,
	}
}

// overrideSpanFromScheduleRow converts a per-schedule override row into an
// overrideSpan.
func overrideSpanFromScheduleRow(row gadb.SchedMgrOverridesForScheduleRow) overrideSpan {
	return overrideSpan{
		AddUserID:    row.AddUserID,
		RemoveUserID: row.RemoveUserID,
		Start:        row.StartTime,
		End:          row.EndTime,
	}
}

// overrideActive reports whether the override o is in effect at now.
//
// The window is INCLUSIVE on both ends (Start <= now <= End), which is exactly
// the `now() BETWEEN start_time AND end_time` filter the batch
// SchedMgrOverrides query applies in SQL. The two MUST stay in agreement: a
// later change to half-open [Start, End) semantics has to change this function
// and the batch query's WHERE clause together.
func overrideActive(now time.Time, o overrideSpan) bool {
	return !now.Before(o.Start) && !now.After(o.End)
}

type updateInfo struct {
	ScheduleID      uuid.UUID
	TimeZone        *time.Location
	RawScheduleData json.RawMessage
	ScheduleData    schedule.Data
	CurrentOnCall   mapset.Set[uuid.UUID]
	Rules           []gadb.SchedMgrRulesRow

	// Overrides are the overrides to consider for this schedule. They may
	// include overrides that are not in effect at the evaluation time; those
	// are skipped by calcLatestOnCall (see overrideActive).
	Overrides []overrideSpan

	// MissingUsers contains user IDs referenced by temporary schedule shifts that no longer exist
	MissingUsers mapset.Set[uuid.UUID]
}

type updateResult struct {
	ScheduleID           uuid.UUID
	UsersToStart         mapset.Set[uuid.UUID]
	UsersToStop          mapset.Set[uuid.UUID]
	NewRawScheduleData   json.RawMessage       // no update necessary if nil
	NotificationChannels mapset.Set[uuid.UUID] // channels to notify, empty if no notifications
}

func (info updateInfo) calcLatestOnCall(now time.Time) mapset.Set[uuid.UUID] {
	if isActive, users := info.ScheduleData.TempOnCall(now); isActive {
		// temporary schedule config takes precedence over anything else and makes this easy
		// we use a thread-unsafe set here to avoid the cost of locking since this is only called in a single thread
		onCall := mapset.NewThreadUnsafeSet(users...)
		if info.MissingUsers != nil {
			// users that no longer exist can never be recorded as on-call, so including them
			// would cause a loop of "on-call changed" state (and repeated notifications)
			onCall.RemoveAll(info.MissingUsers.ToSlice()...)
		}
		return onCall
	}
	now = now.In(info.TimeZone)
	newOnCall := mapset.NewThreadUnsafeSet[uuid.UUID]()
	for _, r := range info.Rules {
		if ruleRowIsActive(r, now) {
			newOnCall.Add(r.ResolvedUserID)
		}
	}

	for _, o := range info.Overrides {
		if !overrideActive(now, o) {
			continue
		}

		if o.RemoveUserID.Valid {
			if !newOnCall.Contains(o.RemoveUserID.UUID) {
				// if the user to be removed is not currently on-call, skip
				// so we don't add a user for replace overrides that don't match
				continue
			}
			newOnCall.Remove(o.RemoveUserID.UUID)
		}

		if o.AddUserID.Valid {
			newOnCall.Add(o.AddUserID.UUID)
		}
	}

	return newOnCall
}

func (info updateInfo) calcUpdates(now time.Time) (*updateResult, error) {
	if info.CurrentOnCall == nil {
		info.CurrentOnCall = mapset.NewThreadUnsafeSet[uuid.UUID]()
	}

	result := updateResult{
		ScheduleID: info.ScheduleID,
		// since we do this in a single thread, we can use a thread-unsafe set and avoid the cost of locking
		UsersToStart:         mapset.NewThreadUnsafeSet[uuid.UUID](),
		UsersToStop:          mapset.NewThreadUnsafeSet[uuid.UUID](),
		NotificationChannels: mapset.NewThreadUnsafeSet[uuid.UUID](),
	}
	now = now.In(info.TimeZone)

	newOnCall := info.calcLatestOnCall(now)
	onCallChanged := !newOnCall.Equal(info.CurrentOnCall)
	if onCallChanged {
		result.UsersToStop = info.CurrentOnCall.Difference(newOnCall)  // currently on-call, but not anymore
		result.UsersToStart = newOnCall.Difference(info.CurrentOnCall) // not currently on-call, but should be
	}

	var dataNeedsUpdate bool
	newRules := make([]schedule.OnCallNotificationRule, len(info.ScheduleData.V1.OnCallNotificationRules))
	// we copy the rules to avoid modifying the original slice
	copy(newRules, info.ScheduleData.V1.OnCallNotificationRules)
	for i, r := range newRules {
		if r.Time == nil { // if time is not set, then it is a "when schedule changes" rule
			if onCallChanged {
				result.NotificationChannels.Add(r.ChannelID)
			}
			continue
		}
		if r.NextNotification != nil && !r.NextNotification.After(now) {
			result.NotificationChannels.Add(r.ChannelID)
		}
		newTime := nextOnCallNotification(now, r)
		if equalTimePtr(r.NextNotification, newTime) {
			// no change, skip
			continue
		}
		dataNeedsUpdate = true
		newRules[i].NextNotification = newTime
	}

	if dataNeedsUpdate {
		info.ScheduleData.V1.OnCallNotificationRules = newRules
		jsonData, err := jsonutil.Apply(info.RawScheduleData, info.ScheduleData)
		if err != nil {
			return nil, fmt.Errorf("apply schedule data: %w", err)
		}
		result.NewRawScheduleData = jsonData
	}

	return &result, nil
}
