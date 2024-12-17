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

		var data schedule.Data
		err = json.Unmarshal(dataRow.Data, &data)
		if err != nil {
			db.logger.ErrorContext(ctx, "failed to unmarshal schedule data, skipping.", slog.String("error", err.Error()), slog.String("schedule_id", dataRow.ScheduleID.String()))

			// Mark as skipped so we don't keep trying to process it.
			return false, gadb.New(tx).CleanupMgrScheduleDataSkip(ctx, dataRow.ScheduleID)
		}

		now, err := gadb.New(tx).Now(ctx)
		if err != nil {
			return false, fmt.Errorf("get current time: %w", err)
		}

		changed, users := trimExpiredShifts(&data, now)
		validUsers, err := gadb.New(tx).CleanupMgrVerifyUsers(ctx, users)
		if err != nil {
			return false, fmt.Errorf("verify users: %w", err)
		}

		changed = trimInvalidUsers(&data, validUsers) || changed
		if !changed {
			return false, gadb.New(tx).CleanupMgrScheduleDataSkip(ctx, dataRow.ScheduleID)
		}

		rawData, err := json.Marshal(data)
		if err != nil {
			return false, fmt.Errorf("marshal schedule data: %w", err)
		}

		db.logger.InfoContext(ctx, "Updated schedule data.", slog.String("schedule_id", dataRow.ScheduleID.String()))
		return false, gadb.New(tx).CleanupMgrUpdateScheduleData(ctx, gadb.CleanupMgrUpdateScheduleDataParams{
			ScheduleID: dataRow.ScheduleID,
			Data:       rawData,
		})
	})
	if err != nil {
		return err
	}

	return nil
}

func trimInvalidUsers(data *schedule.Data, validUsers []uuid.UUID) (changed bool) {
	newTempSched := data.V1.TemporarySchedules[:0]
	for _, temp := range data.V1.TemporarySchedules {
		cleanShifts := temp.Shifts[:0]
		for _, shift := range temp.Shifts {
			id, err := uuid.Parse(shift.UserID)
			if err != nil {
				changed = true
				// invalid shift, delete it
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

// trimExpiredShifts will trim any past shifts and collect all remaining user IDs.
func trimExpiredShifts(data *schedule.Data, cutoff time.Time) (changed bool, users []uuid.UUID) {
	newTempSched := data.V1.TemporarySchedules[:0]
	for _, sched := range data.V1.TemporarySchedules {
		if sched.End.Before(cutoff) {
			changed = true
			continue
		}
		cleanShifts := sched.Shifts[:0]
		for _, shift := range sched.Shifts {
			if shift.End.Before(cutoff) {
				changed = true
				continue
			}
			id, err := uuid.Parse(shift.UserID)
			if err != nil {
				changed = true
				// invalid shift, delete it
				continue
			}

			cleanShifts = append(cleanShifts, shift)
			if slices.Contains(users, id) {
				continue
			}

			users = append(users, id)
		}
		sched.Shifts = cleanShifts
		newTempSched = append(newTempSched, sched)
	}
	data.V1.TemporarySchedules = newTempSched
	return changed, users
}
