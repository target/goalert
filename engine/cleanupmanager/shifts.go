package cleanupmanager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
)

type ShiftArgs struct{}

func (ShiftArgs) Kind() string { return "cleanup-manager-alerts" }

// CleanupShifts will automatically cleanup old shift and override records.
func (db *DB) CleanupShifts(ctx context.Context, j *river.Job[ShiftArgs]) error {
	cfg := config.FromContext(ctx)
	if cfg.Maintenance.ScheduleCleanupDays <= 0 {
		return nil
	}

	err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		count, err := gadb.New(tx).CleanupMgrDeleteOldScheduleShifts(ctx, int64(cfg.Maintenance.ScheduleCleanupDays))
		if err != nil {
			return false, fmt.Errorf("delete old shifts: %w", err)
		}
		return count < 100, nil
	})
	if err != nil {
		return err
	}

	err = db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		count, err := gadb.New(tx).CleanupMgrDeleteOldOverrides(ctx, int64(cfg.Maintenance.ScheduleCleanupDays))
		if err != nil {
			return false, fmt.Errorf("delete old overrides: %w", err)
		}
		return count < 100, nil
	})
	if err != nil {
		return err
	}

	err = db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		count, err := gadb.New(tx).CleanupMgrDeleteOldStepShifts(ctx, int64(cfg.Maintenance.ScheduleCleanupDays))
		if err != nil {
			return false, fmt.Errorf("delete old step shifts: %w", err)
		}
		return count < 100, nil
	})
	if err != nil {
		return err
	}

	return nil
}
