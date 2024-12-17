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
	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/schedule"
)

type SchedDataArgs struct{}

func (SchedDataArgs) Kind() string { return "cleanup-manager-sched-data" }

// CleanupScheduleData will automatically cleanup schedule data.
// - Remove temporary-schedule shifts for users that no longer exist.
// - Remove temporary-schedule shifts that occur in the past.
func (db *DB) CleanupScheduleData(ctx context.Context, j *river.Job[SchedDataArgs]) error {
	cfg := config.FromContext(ctx)
	if cfg.Maintenance.ScheduleCleanupDays <= 0 {
		return nil
	}

	err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		dataRow, err := gadb.New(tx).CleanupMgrScheduleData(ctx)
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("get schedule data: %w", err)
		}
		log := db.logger.With(slog.String("schedule_id", dataRow.ScheduleID.String()))
		gdb := gadb.New(tx)

		var data schedule.Data
		err = json.Unmarshal(dataRow.Data, &data)
		if err != nil {
			log.ErrorContext(ctx,
				"failed to unmarshal schedule data, skipping.",
				slog.String("error", err.Error()))

			// Mark as skipped so we don't keep trying to process it.
			return false, gdb.CleanupMgrScheduleDataSkip(ctx, dataRow.ScheduleID)
		}

		users := collectUsers(data)
		var validUsers []uuid.UUID
		if len(users) > 0 {
			validUsers, err = gdb.CleanupMgrVerifyUsers(ctx, users)
			if err != nil {
				return false, fmt.Errorf("lookup valid users: %w", err)
			}
		}

		now, err := gdb.Now(ctx)
		if err != nil {
			return false, fmt.Errorf("get current time: %w", err)
		}
		changed := cleanupData(&data, validUsers, now)
		if !changed {
			return false, gdb.CleanupMgrScheduleDataSkip(ctx, dataRow.ScheduleID)
		}

		rawData, err := json.Marshal(data)
		if err != nil {
			return false, fmt.Errorf("marshal schedule data: %w", err)
		}

		log.InfoContext(ctx, "Updated schedule data.")
		return false, gdb.CleanupMgrUpdateScheduleData(ctx,
			gadb.CleanupMgrUpdateScheduleDataParams{
				ScheduleID: dataRow.ScheduleID,
				Data:       rawData,
			})
	})
	if err != nil {
		return err
	}

	return nil
}

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
