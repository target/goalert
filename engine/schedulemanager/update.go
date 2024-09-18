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
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/jsonutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
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

	var now time.Time
	err = tx.Stmt(db.currentTime).QueryRowContext(ctx).Scan(&now)
	if err != nil {
		return errors.Wrap(err, "get DB time")
	}

	scheduleData := make(map[string]*schedule.Data)
	rawScheduleData := make(map[string]json.RawMessage)
	rows, err := tx.StmtContext(ctx, db.data).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get schedule data")
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		var data json.RawMessage
		err = rows.Scan(&id, &data)
		if err != nil {
			return errors.Wrap(err, "scan schedule data")
		}
		rawScheduleData[id] = data

		var sData schedule.Data
		err = json.Unmarshal(data, &sData)
		if err != nil {
			log.Log(log.WithField(ctx, "ScheduleID", id), errors.Wrap(err, "unmarshal schedule data "+string(data)))
			continue
		}
		scheduleData[id] = &sData
	}

	rows, err = tx.Stmt(db.overrides).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get active overrides")
	}
	defer rows.Close()

	var overrides []override.UserOverride
	for rows.Next() {
		var o override.UserOverride
		var schedTgt sql.NullString
		var add, rem sql.NullString
		err = rows.Scan(&add, &rem, &schedTgt)
		if err != nil {
			return errors.Wrap(err, "scan override")
		}
		o.AddUserID = add.String
		o.RemoveUserID = rem.String
		if !schedTgt.Valid {
			continue
		}
		o.Target = assignment.ScheduleTarget(schedTgt.String)
		overrides = append(overrides, o)
	}

	rows, err = tx.Stmt(db.rules).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get rules")
	}
	defer rows.Close()

	type userRule struct {
		rule.Rule
		UserID string
	}

	var rules []userRule
	for rows.Next() {
		var r userRule
		err = rows.Scan(
			&r.ScheduleID,
			&r.WeekdayFilter,
			&r.Start,
			&r.End,
			&r.UserID,
		)
		if err != nil {
			return errors.Wrap(err, "scan rule")
		}

		rules = append(rules, r)
	}

	rows, err = tx.StmtContext(ctx, db.schedTZ).QueryContext(ctx)
	if err != nil {
		return fmt.Errorf("fetch schedule TZ info: %w", err)
	}
	defer rows.Close()
	tz := make(map[string]*time.Location)
	for rows.Next() {
		var id, tzName string
		err = rows.Scan(&id, &tzName)
		if err != nil {
			return fmt.Errorf("scan schedule TZ info: %w", err)
		}
		tz[id], err = util.LoadLocation(tzName)
		if err != nil {
			return fmt.Errorf("load TZ info '%s' for schedule '%s': %w", tzName, id, err)
		}
	}

	rows, err = tx.Stmt(db.getOnCall).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "get on call")
	}
	defer rows.Close()

	type onCall struct {
		UserID     string
		ScheduleID string
	}
	oldOnCall := make(map[onCall]bool)
	var oc onCall
	for rows.Next() {
		err = rows.Scan(&oc.ScheduleID, &oc.UserID)
		if err != nil {
			return errors.Wrap(err, "scan on call user")
		}
		oldOnCall[oc] = true
	}

	// Calculate new state
	newOnCall := make(map[onCall]bool, len(rules))

	tempSched := make(map[string]struct{})
	for id, data := range scheduleData {
		ok, users := data.TempOnCall(now)
		if !ok {
			continue
		}

		for _, uid := range users {
			newOnCall[onCall{ScheduleID: id, UserID: uid}] = true
		}
		tempSched[id] = struct{}{}
	}

	for _, r := range rules {
		if _, ok := tempSched[r.ScheduleID]; ok {
			// temp schedule active for this ID, skip
			continue
		}
		if r.IsActive(now.In(tz[r.ScheduleID])) {
			newOnCall[onCall{ScheduleID: r.ScheduleID, UserID: r.UserID}] = true
		}
	}

	for _, o := range overrides {
		if _, ok := tempSched[o.Target.TargetID()]; ok {
			// temp schedule active for this ID, skip
			continue
		}
		if o.AddUserID != "" && o.RemoveUserID == "" {
			// ADD override
			newOnCall[onCall{ScheduleID: o.Target.TargetID(), UserID: o.AddUserID}] = true
			continue
		}
		if o.AddUserID == "" && o.RemoveUserID != "" {
			// REMOVE override
			delete(newOnCall, onCall{ScheduleID: o.Target.TargetID(), UserID: o.RemoveUserID})
			continue
		}

		if newOnCall[onCall{ScheduleID: o.Target.TargetID(), UserID: o.RemoveUserID}] {
			// REPLACE override
			delete(newOnCall, onCall{ScheduleID: o.Target.TargetID(), UserID: o.RemoveUserID})
			newOnCall[onCall{ScheduleID: o.Target.TargetID(), UserID: o.AddUserID}] = true
		}
	}

	start := tx.Stmt(db.startOnCall)

	changedSchedules := make(map[string]struct{})
	for oc := range newOnCall {
		// not on call in DB, but are now
		if !oldOnCall[oc] {
			changedSchedules[oc.ScheduleID] = struct{}{}
			_, err = start.ExecContext(ctx, oc.ScheduleID, oc.UserID)
			if err != nil && !isScheduleDeleted(err) {
				return errors.Wrap(err, "record shift start")
			}
		}
	}
	end := tx.Stmt(db.endOnCall)
	for oc := range oldOnCall {
		// on call in DB, but no longer
		if !newOnCall[oc] {
			changedSchedules[oc.ScheduleID] = struct{}{}
			_, err = end.ExecContext(ctx, oc.ScheduleID, oc.UserID)
			if err != nil {
				return errors.Wrap(err, "record shift end")
			}
		}
	}

	// Notify changed schedules
	needsOnCallNotification := make(map[string][]uuid.UUID)
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
		_, err = tx.StmtContext(ctx, db.updateData).ExecContext(ctx, schedID, jsonData)
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
			_, err = tx.StmtContext(ctx, db.scheduleOnCallNotification).ExecContext(ctx, uuid.New(), chanID, schedID)
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
