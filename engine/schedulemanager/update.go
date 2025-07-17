package schedulemanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
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
			ScheduleID: schedID,
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
		info.Overrides = append(info.Overrides, o)
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

	onCallRows, err := q.SchedMgrOnCall(ctx)
	if err != nil {
		return errors.Wrap(err, "get on call")
	}

	for _, row := range onCallRows {
		getInfo(row.ScheduleID).CurrentOnCall.Add(row.UserID)
	}

	for scheduleID, info := range updateData {
		result, err := info.calcUpdates(now)
		if err != nil {
			log.Log(log.WithField(ctx, "ScheduleID", scheduleID), errors.Wrap(err, "calc updates"))
			continue
		}

		for userID := range result.UsersToStart.Each {
			err = q.SchedMgrStartOnCall(ctx, gadb.SchedMgrStartOnCallParams{
				ScheduleID: info.ScheduleID,
				UserID:     userID,
			})
			if err != nil {
				return errors.Wrapf(err, "record shift start for user %s on schedule %s", userID, info.ScheduleID)
			}
		}
		for userID := range result.UsersToStop.Each {
			err = q.SchedMgrEndOnCall(ctx, gadb.SchedMgrEndOnCallParams{
				ScheduleID: info.ScheduleID,
				UserID:     userID,
			})
			if err != nil {
				return errors.Wrapf(err, "record shift end for user %s on schedule %s", userID, info.ScheduleID)
			}
		}

		if result.NewRawScheduleData != nil {
			err = q.SchedMgrSetData(ctx, gadb.SchedMgrSetDataParams{
				ScheduleID: info.ScheduleID,
				Data:       result.NewRawScheduleData,
			})
			if err != nil {
				return errors.Wrapf(err, "set schedule data for %s", info.ScheduleID)
			}
		}

		for chanID := range result.NotificationChannels.Each {
			err = q.SchedMgrInsertMessage(ctx, gadb.SchedMgrInsertMessageParams{
				ID:         uuid.New(),
				ChannelID:  uuid.NullUUID{UUID: chanID, Valid: true},
				ScheduleID: uuid.NullUUID{UUID: info.ScheduleID, Valid: true},
			})
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
