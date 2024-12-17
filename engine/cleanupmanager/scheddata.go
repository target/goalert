package cleanupmanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
)

type SchedDataArgs struct {
	ScheduleID uuid.UUID
}

func (SchedDataArgs) Kind() string { return "cleanup-manager-sched-data" }

type SchedDataLookForWorkArgs struct{}

func (SchedDataLookForWorkArgs) Kind() string { return "cleanup-manager-sched-data-look-for-work" }

// LookForWorkScheduleData will automatically look for schedules that need their JSON data cleaned up and insert them into the queue.
func (db *DB) LookForWorkScheduleData(ctx context.Context, j *river.Job[SchedDataLookForWorkArgs]) error {
	cfg := config.FromContext(ctx)
	if cfg.Maintenance.ScheduleCleanupDays <= 0 {
		return nil
	}
	var outOfDate []uuid.UUID
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		// Grab schedules that haven't been cleaned up in the last 30 days.
		outOfDate, err = gadb.New(tx).CleanupMgrScheduleNeedsCleanup(ctx, 30)
		return err
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	var params []river.InsertManyParams
	for _, id := range outOfDate {
		params = append(params, river.InsertManyParams{
			Args: SchedDataArgs{ScheduleID: id},
			InsertOpts: &river.InsertOpts{
				Queue:      QueueName,
				Priority:   PriorityTempSched,
				UniqueOpts: river.UniqueOpts{ByArgs: true},
			},
		})
	}

	if len(params) == 0 {
		return nil
	}

	_, err = river.ClientFromContext[pgx.Tx](ctx).InsertMany(ctx, params)
	if err != nil {
		return fmt.Errorf("insert many: %w", err)
	}

	return nil
}

// CleanupScheduleData will automatically cleanup schedule data.
// - Remove temporary-schedule shifts for users that no longer exist.
// - Remove temporary-schedule shifts that occur in the past.
func (db *DB) CleanupScheduleData(ctx context.Context, j *river.Job[SchedDataArgs]) error {
	cfg := config.FromContext(ctx)
	if cfg.Maintenance.ScheduleCleanupDays <= 0 {
		return nil
	}
	log := db.logger.With(slog.String("schedule_id", j.Args.ScheduleID.String()))

	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) (err error) {
		// Grab the next schedule that hasn't been cleaned up in the last 30 days.
		rawData, err := gadb.New(tx).CleanupMgrScheduleData(ctx, j.Args.ScheduleID)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("get schedule data: %w", err)
		}
		gdb := gadb.New(tx)

		var data schedule.Data
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return fmt.Errorf("unmarshal schedule data: %w", err)
		}

		// We want to remove shifts for users that no longer exist, so to do that we'll get the set of users from the schedule data and verify them.
		users := collectUsers(data)
		var validUsers []uuid.UUID
		if len(users) > 0 {
			validUsers, err = gdb.CleanupMgrVerifyUsers(ctx, users)
			if err != nil {
				return fmt.Errorf("lookup valid users: %w", err)
			}
		}

		now, err := gdb.Now(ctx)
		if err != nil {
			return fmt.Errorf("get current time: %w", err)
		}
		changed := cleanupData(&data, validUsers, now)
		if !changed {
			return gdb.CleanupMgrScheduleDataSkip(ctx, j.Args.ScheduleID)
		}

		rawData, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal schedule data: %w", err)
		}

		log.InfoContext(ctx, "Updated schedule data.")
		return gdb.CleanupMgrUpdateScheduleData(ctx,
			gadb.CleanupMgrUpdateScheduleDataParams{
				ScheduleID: j.Args.ScheduleID,
				Data:       rawData,
			})
	})
	if err != nil {
		return err
	}

	return nil
}

// cleanupData will cleanup the schedule data, removing temporary-schedules that occur in the past; and removing shifts for users that no longer exist.
func cleanupData(data *schedule.Data, validUsers []uuid.UUID, now time.Time) (changed bool) {
	newTempSched := data.V1.TemporarySchedules[:0]
	for _, temp := range data.V1.TemporarySchedules {
		if temp.End.Before(now) {
			changed = true
			continue
		}

		cleanShifts := temp.Shifts[:0]
		for _, shift := range temp.Shifts {
			id, err := uuid.Parse(shift.UserID)
			if err != nil {
				changed = true
				// invalid user id/shift, delete it
				continue
			}
			if !slices.Contains(validUsers, id) {
				changed = true
				continue
			}
			cleanShifts = append(cleanShifts, shift)
		}
		temp.Shifts = cleanShifts
		newTempSched = append(newTempSched, temp)
	}
	data.V1.TemporarySchedules = newTempSched
	return changed
}

// collectUsers will collect all user ids from the schedule data.
func collectUsers(data schedule.Data) (users []uuid.UUID) {
	for _, sched := range data.V1.TemporarySchedules {
		for _, shift := range sched.Shifts {
			id, err := uuid.Parse(shift.UserID)
			if err != nil {
				// invalid id, skip it
				continue
			}

			if slices.Contains(users, id) {
				continue
			}

			users = append(users, id)
		}
	}

	return users
}
