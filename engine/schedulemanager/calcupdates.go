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

type updateInfo struct {
	ScheduleID      uuid.UUID
	TimeZone        *time.Location
	RawScheduleData json.RawMessage
	ScheduleData    schedule.Data
	CurrentOnCall   mapset.Set[uuid.UUID]
	Rules           []gadb.SchedMgrRulesRow
	Overrides       []gadb.UserOverride
}

type updateResult struct {
	ScheduleID           uuid.UUID `json:"-"`
	UsersToStart         mapset.Set[uuid.UUID]
	UsersToStop          mapset.Set[uuid.UUID]
	NewRawScheduleData   json.RawMessage       // no update necessary if nil
	NotificationChannels mapset.Set[uuid.UUID] // channels to notify, empty if no notifications
}

func (info updateInfo) calcLatestOnCall(now time.Time) mapset.Set[uuid.UUID] {
	if isActive, users := info.ScheduleData.TempOnCall(now); isActive {
		// temporary schedule config takes precedence over anything else and makes this easy
		// we use a thread-unsafe set here to avoid the cost of locking since this is only called in a single thread
		return mapset.NewThreadUnsafeSet(users...)
	}
	now = now.In(info.TimeZone)
	newOnCall := mapset.NewThreadUnsafeSet[uuid.UUID]()
	for _, r := range info.Rules {
		if ruleRowIsActive(r, now) {
			newOnCall.Add(r.ResolvedUserID)
		}
	}

	for _, o := range info.Overrides {
		if now.Before(o.StartTime) || !now.Before(o.EndTime) {
			// not active
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
