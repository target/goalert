package cleanupmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util"
)

type AlertArgs struct{}

func (AlertArgs) Kind() string { return "cleanup-manager-alerts" }

// CleanupAlerts will automatically close and delete old alerts.
func (db *DB) CleanupAlerts(ctx context.Context, j *river.Job[AlertArgs]) error {
	cfg := config.FromContext(ctx)

	for cfg.Maintenance.AlertAutoCloseDays > 0 {
		var count int64
		err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
			ids, err := gadb.New(tx).CleanupMgrFindStaleAlerts(ctx, gadb.CleanupMgrFindStaleAlertsParams{
				AutoCloseDays: int64(cfg.Maintenance.AlertAutoCloseDays),
				IncludeAcked:  cfg.Maintenance.AutoCloseAckedAlerts,
			})
			if err != nil {
				return fmt.Errorf("find stale alerts: %w", err)
			}

			var idsInt []int
			for _, id := range ids {
				idsInt = append(idsInt, int(id))
			}

			_, err = db.alertStore.UpdateManyAlertStatus(ctx, alert.StatusClosed, idsInt, alertlog.AutoClose{AlertAutoCloseDays: cfg.Maintenance.AlertAutoCloseDays})
			if err != nil {
				return fmt.Errorf("update alerts: %w", err)
			}

			count = int64(len(ids))
			return nil
		})
		if err != nil {
			return fmt.Errorf("auto close alerts: %w", err)
		}
		if count < 100 {
			// Assume we've closed all old alerts, since we got less than 100 (which is the max we close at once).
			break
		}

		err = util.ContextSleep(ctx, 100*time.Millisecond)
		if err != nil {
			return fmt.Errorf("auto close alerts: sleep: %w", err)
		}
	}

	for cfg.Maintenance.AlertCleanupDays > 0 {
		var count int64
		err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
			var err error
			count, err = gadb.New(tx).CleanupMgrDeleteOldAlerts(ctx, int64(cfg.Maintenance.AlertCleanupDays))
			return err
		})
		if err != nil {
			return fmt.Errorf("delete old alerts: %w", err)
		}
		if count < 100 {
			// Assume we've deleted all old alerts, since we got less than 100 (which is the max we delete at once).
			break
		}

		err = util.ContextSleep(ctx, 100*time.Millisecond)
		if err != nil {
			return fmt.Errorf("delete old alerts: sleep: %w", err)
		}
	}

	return nil
}
