package cleanupmanager

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

// UpdateAll will update the state of all active escalation policies.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := db.update(ctx)
	return err
}

func (db *DB) update(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running cleanup operations.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.StmtContext(ctx, db.setTimeout).ExecContext(ctx)
	if err != nil {
		return err
	}

	cfg := config.FromContext(ctx)
	if cfg.Maintenance.AlertCleanupDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.AlertCleanupDays)
		dur.Status = pgtype.Present
		_, err = db.cleanupAlerts.ExecContext(ctx, &dur)
		if err != nil {
			return err
		}
	}
	if cfg.Maintenance.APIKeyExpireDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.APIKeyExpireDays)
		dur.Status = pgtype.Present
		_, err = db.cleanupAPIKeys.ExecContext(ctx, &dur)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
