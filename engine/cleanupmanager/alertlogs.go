package cleanupmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

type AlertLogArgs struct{}

func (AlertLogArgs) Kind() string { return "cleanup-manager-alert-logs" }

type timeoutWorker[T river.JobArgs] struct {
	river.Worker[T]
	timeout time.Duration
}

// Timeout implements Worker interface.
func (w *timeoutWorker[T]) Timeout(job *river.Job[T]) time.Duration {
	return w.timeout
}

// CleanupAlertLogs will remove alert log entries for deleted alerts.
func (db *DB) CleanupAlertLogs(ctx context.Context, j *river.Job[AlertLogArgs]) error {
	var lastID int64 // start at zero, we will scan _all_ logs

	err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		db.logger.DebugContext(ctx, "Cleaning up alert logs...", "lastID", lastID)
		lastID, err = gadb.New(tx).CleanupAlertLogs(ctx,
			gadb.CleanupAlertLogsParams{
				BatchSize: 10000,
				AfterID:   lastID,
			})
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		if err != nil {
			return false, fmt.Errorf("cleanup alert logs: %w", err)
		}

		return false, nil
	})
	if err != nil {
		return fmt.Errorf("cleanup alert logs: %w", err)
	}

	return nil
}
