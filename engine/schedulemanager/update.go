package schedulemanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util"
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

func ruleRowIsActive(row gadb.SchedMgrRulesRow, t time.Time) bool {
	return ruleFromRow(row).IsActive(t)
}

func (db *DB) update(ctx context.Context) error {
	tx, err := db.lock.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer sqlutil.Rollback(ctx, "schedule manager", tx)

	log.Debugf(ctx, "Updating schedule rules.")

	q := gadb.New(tx)
	now, err := q.Now(ctx)
	if err != nil {
		return errors.Wrap(err, "get DB time")
	}

	updateData := make(map[uuid.UUID]*updateInfo)
	getInfo := func(schedID uuid.UUID) *updateInfo {
		if info, ok := updateData[schedID]; ok {
			return info
		}
		info := &updateInfo{
			ScheduleID:    schedID,
			CurrentOnCall: mapset.NewThreadUnsafeSet[uuid.UUID](),
		}
		updateData[schedID] = info
		return info
	}

	dataRows, err := q.SchedMgrDataForUpdate(ctx)
	if err != nil {
		return errors.Wrap(err, "get schedule data")
	}

	for _, row := range dataRows {
		info := getInfo(row.ScheduleID)
		info.RawScheduleData = row.Data

		var sData schedule.Data
		err = json.Unmarshal(row.Data, &sData)
		if err != nil {
			log.Log(log.WithField(ctx, "ScheduleID", row.ScheduleID), errors.Wrap(err, "unmarshal schedule data "+string(row.Data)))
			continue
		}
		info.ScheduleData = sData
	}

	overrides, err := q.SchedMgrOverrides(ctx)
	if err != nil {
		return errors.Wrap(err, "get active overrides")
	}
	for _, o := range overrides {
		info := getInfo(o.TgtScheduleID)
		info.ActiveOverrides = append(info.ActiveOverrides, o)
	}

	rules, err := q.SchedMgrRules(ctx)
	if err != nil {
		return errors.Wrap(err, "get rules")
	}
	for _, r := range rules {
		info := getInfo(r.ScheduleID)
		info.Rules = append(info.Rules, r)
	}

	tzRows, err := q.SchedMgrTimezones(ctx)
	if err != nil {
		return fmt.Errorf("get timezones: %w", err)
	}
	for _, row := range tzRows {
		getInfo(row.ID).TimeZone, err = util.LoadLocation(row.TimeZone)
		if err != nil {
			return fmt.Errorf("load TZ info '%s' for schedule '%s': %w", row.TimeZone, row.ID, err)
		}
	}

	// Temporary schedule shifts are stored as JSON, so they may reference users that have since been deleted.
	tempUsers := mapset.NewThreadUnsafeSet[uuid.UUID]()
	for _, info := range updateData {
		if isActive, users := info.ScheduleData.TempOnCall(now); isActive {
			tempUsers.Append(users...)
		}
	}
	if tempUsers.Cardinality() > 0 {
		validUsers, err := q.SchedMgrVerifyUsers(ctx, tempUsers.ToSlice())
		if err != nil {
			return errors.Wrap(err, "verify temporary schedule users")
		}
		missing := tempUsers.Clone()
		missing.RemoveAll(validUsers...)
		if missing.Cardinality() > 0 {
			for _, info := range updateData {
				info.MissingUsers = missing
			}
		}
	}

	onCallRows, err := q.SchedMgrOnCall(ctx)
	if err != nil {
		return errors.Wrap(err, "get on call")
	}

	for _, row := range onCallRows {
		getInfo(row.ScheduleID).CurrentOnCall.Add(row.UserID)
	}

updateLoop:
	for scheduleID, info := range updateData {
		result, err := info.calcUpdates(now)
		if err != nil {
			log.Log(log.WithField(ctx, "ScheduleID", scheduleID), errors.Wrap(err, "calc updates"))
			continue
		}

		for userID := range mapset.Elements(result.UsersToStart) {
			err = q.SchedMgrStartOnCall(ctx, gadb.SchedMgrStartOnCallParams{
				ScheduleID: info.ScheduleID,
				UserID:     userID,
			})
			if isScheduleDeleted(err) {
				// If the schedule was deleted, we skip this update and continue with the next one.
				// This prevents errors from being returned when the schedule no longer exists.
				continue updateLoop
			}
			if err != nil {
				return errors.Wrapf(err, "record shift start for user %s on schedule %s", userID, info.ScheduleID)
			}
		}
		for userID := range mapset.Elements(result.UsersToStop) {
			err = q.SchedMgrEndOnCall(ctx, gadb.SchedMgrEndOnCallParams{
				ScheduleID: info.ScheduleID,
				UserID:     userID,
			})
			if isScheduleDeleted(err) {
				continue updateLoop
			}
			if err != nil {
				return errors.Wrapf(err, "record shift end for user %s on schedule %s", userID, info.ScheduleID)
			}
		}

		if result.NewRawScheduleData != nil {
			err = q.SchedMgrSetData(ctx, gadb.SchedMgrSetDataParams{
				ScheduleID: info.ScheduleID,
				Data:       result.NewRawScheduleData,
			})
			if isScheduleDeleted(err) {
				continue
			}
			if err != nil {
				return errors.Wrapf(err, "set schedule data for %s", info.ScheduleID)
			}
		}

		for chanID := range mapset.Elements(result.NotificationChannels) {
			err = q.SchedMgrInsertMessage(ctx, gadb.SchedMgrInsertMessageParams{
				ID:         uuid.New(),
				ChannelID:  uuid.NullUUID{UUID: chanID, Valid: true},
				ScheduleID: uuid.NullUUID{UUID: info.ScheduleID, Valid: true},
			})
			if isScheduleDeleted(err) {
				continue
			}
			if err != nil {
				return errors.Wrapf(err, "insert notification message for channel %s on schedule %s", chanID, info.ScheduleID)
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
			newTime = rule.Time.FirstOfDay(timeutil.AddDays(nowInZone, 1))
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
	if err == nil {
		return false
	}
	dbErr := sqlutil.MapError(err)
	if dbErr == nil {
		return false
	}
	switch dbErr.ConstraintName {
	case "schedule_on_call_users_schedule_id_fkey",
		"schedule_data_schedule_id_fkey",
		"outgoing_messages_schedule_id_fkey":
		return true
	default:
		return false
	}
}
