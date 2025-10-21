package schedulemanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/util/timeutil"
)

type UpdateArgs struct {
	ScheduleID uuid.UUID
}

func (UpdateArgs) Kind() string { return "schedule-manager-update" }

// updateSchedule updates the state of a single schedule, and creates a job for the next change time.
func (db *DB) updateSchedule(ctx context.Context, j *river.Job[UpdateArgs]) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		g := gadb.New(tx)

		now, err := g.Now(ctx)
		if err != nil {
			return fmt.Errorf("get now: %w", err)
		}

		info, err := getUpdateInfo(ctx, tx, j.Args.ScheduleID)
		if isScheduleDeleted(err) {
			// schedule was deleted, nothing to do
			return nil
		}
		if err != nil {
			return fmt.Errorf("get update info: %w", err)
		}

		updates, err := info.calcUpdates(now)
		if err != nil {
			return fmt.Errorf("calculate updates: %w", err)
		}

		if updates.NewRawScheduleData != nil {
			err = g.SchedUpdateData(ctx, gadb.SchedUpdateDataParams{
				ScheduleID: j.Args.ScheduleID,
				Data:       updates.NewRawScheduleData,
			})
			if isScheduleDeleted(err) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("update schedule data: %w", err)
			}
		}

		startUsers := updates.UsersToStart.ToSlice()
		if len(startUsers) > 0 {
			err = g.SchedMgrStartOnCall(ctx, gadb.SchedMgrStartOnCallParams{
				ScheduleID: j.Args.ScheduleID,
				UserIds:    startUsers,
			})
			if isScheduleDeleted(err) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("start on-call: %w", err)
			}
		}

		stopUsers := updates.UsersToStop.ToSlice()
		if len(stopUsers) > 0 {
			err = g.SchedMgrEndOnCall(ctx, gadb.SchedMgrEndOnCallParams{
				ScheduleID: j.Args.ScheduleID,
				UserIds:    stopUsers,
			})
			if isScheduleDeleted(err) {
				return nil
			}
			if err != nil {
				return fmt.Errorf("end on-call: %w", err)
			}
		}

		for chanID := range mapset.Elements(updates.NotificationChannels) {
			err = g.SchedMgrInsertMessage(ctx, gadb.SchedMgrInsertMessageParams{
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

		nextTime := info.nextUpdateTime(now)
		_, err = db.riverDBSQL.InsertTx(ctx, tx, j.Args, &river.InsertOpts{
			UniqueOpts: river.UniqueOpts{
				ByArgs:   true,
				ByPeriod: time.Minute,
			},
			Priority:    PriorityScheduled,
			ScheduledAt: nextTime,
			Queue:       QueueName,
		})
		if err != nil {
			return fmt.Errorf("schedule next run: %w", err)
		}

		return nil
	})
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

func ruleRowIsActive(row gadb.SchedMgrRulesRow, t time.Time) bool {
	var wf timeutil.WeekdayFilter
	if row.ScheduleRule.Sunday {
		wf[0] = 1
	}
	if row.ScheduleRule.Monday {
		wf[1] = 1
	}
	if row.ScheduleRule.Tuesday {
		wf[2] = 1
	}
	if row.ScheduleRule.Wednesday {
		wf[3] = 1
	}
	if row.ScheduleRule.Thursday {
		wf[4] = 1
	}
	if row.ScheduleRule.Friday {
		wf[5] = 1
	}
	if row.ScheduleRule.Saturday {
		wf[6] = 1
	}
	return rule.Rule{
		Start:         row.ScheduleRule.StartTime,
		End:           row.ScheduleRule.EndTime,
		WeekdayFilter: wf,
	}.IsActive(t)
}
