package cleanupmanager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
)

type AlertArgs struct{}

func (AlertArgs) Kind() string { return "cleanup-manager-alerts" }

// CleanupAlerts will automatically close and delete old alerts.
func (db *DB) CleanupAlerts(ctx context.Context, j *river.Job[AlertArgs]) error {
	cfg := config.FromContext(ctx)
	if cfg.Maintenance.AlertAutoCloseDays > 0 {
		err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
			ids, err := gadb.New(tx).CleanupMgrFindStaleAlerts(ctx, gadb.CleanupMgrFindStaleAlertsParams{
				AutoCloseThresholdDays: int64(cfg.Maintenance.AlertAutoCloseDays),
				IncludeAcked:           cfg.Maintenance.AutoCloseAckedAlerts,
			})
			if err != nil {
				return false, fmt.Errorf("find stale alerts: %w", err)
			}

			var idsInt []int
			for _, id := range ids {
				idsInt = append(idsInt, int(id))
			}

			_, err = db.alertStore.UpdateManyAlertStatus(ctx, alert.StatusClosed, idsInt, alertlog.AutoClose{AlertAutoCloseDays: cfg.Maintenance.AlertAutoCloseDays})
			if err != nil {
				return false, fmt.Errorf("update alerts: %w", err)
			}

			return len(ids) < 100, nil
		})
		if err != nil {
			return fmt.Errorf("auto close alerts: %w", err)
		}
	}

	if cfg.Maintenance.AlertCleanupDays > 0 {
		err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
			count, err := gadb.New(tx).CleanupMgrDeleteOldAlerts(ctx, int64(cfg.Maintenance.AlertCleanupDays))
			if err != nil {
				return false, fmt.Errorf("delete old alerts: %w", err)
			}
			return count < 100, nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}
