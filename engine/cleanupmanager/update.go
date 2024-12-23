package cleanupmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
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
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "cleanup manager", tx)

	_, err = tx.StmtContext(ctx, db.setTimeout).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("set timeout: %w", err)
	}

	_, err = tx.StmtContext(ctx, db.cleanupSessions).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("cleanup sessions: %w", err)
	}

	cfg := config.FromContext(ctx)

	if cfg.Maintenance.APIKeyExpireDays > 0 {
		var dur pgtype.Interval
		dur.Days = int32(cfg.Maintenance.APIKeyExpireDays)
		dur.Status = pgtype.Present
		_, err = tx.StmtContext(ctx, db.cleanupAPIKeys).ExecContext(ctx, &dur)
		if err != nil {
			return err
		}
	}

	err = tx.StmtContext(ctx, db.cleanupAlertLogs).QueryRowContext(ctx, db.logIndex).Scan(&db.logIndex)
	if errors.Is(err, sql.ErrNoRows) {
		// repeat
		db.logIndex = 0
		err = nil
	}
	if err != nil {
		return fmt.Errorf("cleanup alert_logs: %w", err)
	}

	return tx.Commit()
}
