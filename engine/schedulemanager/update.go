package schedulemanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/jsonutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/util/timeutil"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = db.update(ctx)
	return err
}

func ruleRowIsActive(row gadb.SchedMgrRulesRow, t time.Time) bool {
	var wf timeutil.WeekdayFilter
	if row.Sunday {
		wf[0] = 1
	}
	if row.Monday {
		wf[1] = 1
	}
	if row.Tuesday {
		wf[2] = 1
	}
	if row.Wednesday {
		wf[3] = 1
	}
	if row.Thursday {
		wf[4] = 1
	}
	if row.Friday {
		wf[5] = 1
	}
	if row.Saturday {
		wf[6] = 1
	}
	return rule.Rule{
		Start:         row.StartTime,
		End:           row.EndTime,
		WeekdayFilter: wf,
	}.IsActive(t)
}

func (db *DB) update(ctx context.Context) error {
	tx, state, err := db.lock.BeginTxWithState(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer sqlutil.Rollback(ctx, "schedule manager", tx)

	var s State
	err = state.Load(ctx, &s)
	if err != nil {
		return errors.Wrap(err, "load state")
	}

	if !s.HasMigratedScheduleData {
		isDone, err := db.migrateScheduleDataNotifDedup(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "migrate schedule data")
		}
		if !isDone {
			// We're not done yet, so we'll try again later.
			return tx.Commit()
		}

		s.HasMigratedScheduleData = true
		err = state.Save(ctx, &s)
		if err != nil {
			return errors.Wrap(err, "save state")
		}
	}

	log.Debugf(ctx, "Updating schedule rules.")

	q := gadb.NewCompat(tx)
	now, err := q.Now(ctx)
	if err != nil {
		return errors.Wrap(err, "get DB time")
	}

	scheduleData := make(map[uuid.UUID]*schedule.Data)
	rawScheduleData := make(map[uuid.UUID]json.RawMessage)
	dataRows, err := q.SchedMgrDataForUpdate(ctx)
	if err != nil {
		return errors.Wrap(err, "get schedule data")
	}

	for _, row := range dataRows {
		rawScheduleData[row.ScheduleID] = row.Data

		var sData schedule.Data
		err = json.Unmarshal(row.Data, &sData)
		if err != nil {
			log.Log(log.WithField(ctx, "ScheduleID", row.ScheduleID), errors.Wrap(err, "unmarshal schedule data "+string(row.Data)))
			continue
		}
		scheduleData[row.ScheduleID] = &sData
	}

	overrides, err := q.SchedMgrOverrides(ctx)
	if err != nil {
		return errors.Wrap(err, "get active overrides")
	}

	rules, err := q.SchedMgrRules(ctx)
	if err != nil {
		return errors.Wrap(err, "get rules")
	}

	tzRows, err := q.SchedMgrTimezones(ctx)
	if err != nil {
		return fmt.Errorf("get timezones: %w", err)
	}
	tz := make(map[uuid.UUID]*time.Location)
	for _, row := range tzRows {
		tz[row.ID], err = util.LoadLocation(row.TimeZone)
		if err != nil {
			return fmt.Errorf("load TZ info '%s' for schedule '%s': %w", row.TimeZone, row.ID, err)
		}
	}

	onCallRows, err := q.SchedMgrOnCall(ctx)
	if err != nil {
		return errors.Wrap(err, "get on call")
	}

	oldOnCall := make(map[gadb.SchedMgrOnCallRow]bool)
	for _, row := range onCallRows {
		oldOnCall[row] = true
	}

	// Calculate new state
	newOnCall := make(map[gadb.SchedMgrOnCallRow]bool, len(rules))

	tempSched := make(map[uuid.UUID]struct{})
	for id, data := range scheduleData {
		ok, users := data.TempOnCall(now)
		if !ok {
			continue
		}

		for _, uid := range users {
			newOnCall[gadb.SchedMgrOnCallRow{ScheduleID: id, UserID: uid}] = true
		}
		tempSched[id] = struct{}{}
	}

	for _, r := range rules {
		if _, ok := tempSched[r.ScheduleID]; ok {
			// temp schedule active for this ID, skip
			continue
		}
		if ruleRowIsActive(r, now.In(tz[r.ScheduleID])) {
			newOnCall[gadb.SchedMgrOnCallRow{ScheduleID: r.ScheduleID, UserID: r.ResolvedUserID}] = true
		}
	}

	for _, o := range overrides {
		if _, ok := tempSched[o.TgtScheduleID]; ok {
			// temp schedule active for this ID, skip
			continue
		}
		if o.RemoveUserID.Valid {
			delete(newOnCall, gadb.SchedMgrOnCallRow{ScheduleID: o.TgtScheduleID, UserID: o.RemoveUserID.UUID})
		}
		if o.AddUserID.Valid {
			newOnCall[gadb.SchedMgrOnCallRow{ScheduleID: o.TgtScheduleID, UserID: o.AddUserID.UUID}] = true
		}
	}

	changedSchedules := make(map[uuid.UUID]struct{})
	for oc := range newOnCall {
		// not on call in DB, but are now
		if !oldOnCall[oc] {
			changedSchedules[oc.ScheduleID] = struct{}{}
			err = q.SchedMgrStartOnCall(ctx, gadb.SchedMgrStartOnCallParams(oc))
			if err != nil && !isScheduleDeleted(err) {
				return errors.Wrap(err, "record shift start")
			}
		}
	}

	for oc := range oldOnCall {
		// on call in DB, but no longer
		if !newOnCall[oc] {
			changedSchedules[oc.ScheduleID] = struct{}{}
			err = q.SchedMgrEndOnCall(ctx, gadb.SchedMgrEndOnCallParams(oc))
			if err != nil {
				return errors.Wrap(err, "record shift end")
			}
		}
	}

	// Notify changed schedules
	needsOnCallNotification := make(map[uuid.UUID][]uuid.UUID)
	for schedID := range changedSchedules {
		data := scheduleData[schedID]
		if data == nil {
			continue
		}
		for _, r := range data.V1.OnCallNotificationRules {
			if r.Time != nil {
				continue
			}

			needsOnCallNotification[schedID] = append(needsOnCallNotification[schedID], r.ChannelID)
		}
	}

	for schedID, data := range scheduleData {
		var hadChange bool
		for i, r := range data.V1.OnCallNotificationRules {
			if r.NextNotification != nil && !r.NextNotification.After(now) {
				needsOnCallNotification[schedID] = append(needsOnCallNotification[schedID], r.ChannelID)
			}

			newTime := nextOnCallNotification(now.In(tz[schedID]), r)
			if equalTimePtr(r.NextNotification, newTime) {
				continue
			}
			hadChange = true
			data.V1.OnCallNotificationRules[i].NextNotification = newTime
		}
		if !hadChange {
			continue
		}

		jsonData, err := jsonutil.Apply(rawScheduleData[schedID], data)
		if err != nil {
			return err
		}
		err = q.SchedMgrSetData(ctx, gadb.SchedMgrSetDataParams{
			ScheduleID: schedID,
			Data:       jsonData,
		})
		if err != nil {
			return err
		}
	}

	for schedID, chanIDs := range needsOnCallNotification {
		sort.Slice(chanIDs, func(i, j int) bool { return chanIDs[i].String() < chanIDs[j].String() })
		var lastID uuid.UUID
		for _, chanID := range chanIDs {
			if chanID == lastID {
				continue
			}
			lastID = chanID
			err = q.SchedMgrInsertMessage(ctx, gadb.SchedMgrInsertMessageParams{
				ID:         uuid.New(),
				ChannelID:  uuid.NullUUID{UUID: chanID, Valid: true},
				ScheduleID: uuid.NullUUID{UUID: schedID, Valid: true},
			})
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func equalTimePtr(a, b *time.Time) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}

	return a.Equal(*b)
}

func nextOnCallNotification(nowInZone time.Time, rule schedule.OnCallNotificationRule) *time.Time {
	if rule.Time == nil {
		return nil
	}
	if rule.WeekdayFilter == nil || rule.WeekdayFilter.IsAlways() {
		newTime := rule.Time.FirstOfDay(nowInZone)
		if !newTime.After(nowInZone) {
			// add a day
			y, m, d := nowInZone.Date()
			nowInZone = time.Date(y, m, d+1, 0, 0, 0, 0, nowInZone.Location())
			newTime = rule.Time.FirstOfDay(nowInZone)
		}

		return &newTime
	}

	if rule.WeekdayFilter.IsNever() {
		return nil
	}

	var newTime time.Time
	if rule.WeekdayFilter.Day(nowInZone.Weekday()) {
		newTime = rule.Time.FirstOfDay(nowInZone)
	} else {
		newTime = rule.Time.FirstOfDay(rule.WeekdayFilter.NextActive(nowInZone))
	}
	if !newTime.After(nowInZone) {
		newTime = rule.Time.FirstOfDay(rule.WeekdayFilter.NextActive(newTime))
	}

	return &newTime
}

func isScheduleDeleted(err error) bool {
	dbErr := sqlutil.MapError(err)
	if dbErr == nil {
		return false
	}
	return dbErr.ConstraintName == "schedule_on_call_users_schedule_id_fkey"
}
