package cleanupmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

type AlertLogLFWArgs struct{}

func (AlertLogLFWArgs) Kind() string { return "cleanup-manager-alert-logs-lfw" }

type AlertLogArgs struct {
	StartID int64
	EndID   int64
}

const (
	batchSize = 5000
	blockSize = 100000
)

// LookForWorkAlertLogs will schedule alert log cleanup jobs for blocks of alert log IDs.
//
// The strategy here is to look for the minimum and maximum alert log IDs in the database, then schedule jobs for each `blockSize` block of IDs,
// and those jobs will then cleanup the alert logs in that range `batchSize` at a time.
func (db *DB) LookForWorkAlertLogs(ctx context.Context, j *river.Job[AlertLogLFWArgs]) error {
	var min, max int64
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		row, err := gadb.New(tx).CleanupMgrAlertLogsMinMax(ctx)
		if err != nil {
			return err
		}
		min, max = row.MinID, row.MaxID
		return nil
	})
	if min == 0 && max == 0 {
		return nil
	}
	if err != nil {
		return fmt.Errorf("get min/max alert log ID: %w", err)
	}

	max++

	var params []river.InsertManyParams
	for i := int64(0); i < max; i += blockSize {
		if i < min {
			// skip sparse blocks
			continue
		}

		params = append(params, river.InsertManyParams{
			Args: AlertLogArgs{StartID: i, EndID: i + blockSize},
			InsertOpts: &river.InsertOpts{
				Queue:      QueueName,
				Priority:   PriorityAlertLogs,
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

func (AlertLogArgs) Kind() string { return "cleanup-manager-alert-logs" }

// CleanupAlertLogs will remove alert log entries for deleted alerts.
func (db *DB) CleanupAlertLogs(ctx context.Context, j *river.Job[AlertLogArgs]) error {
	lastID := j.Args.StartID

	err := db.whileWork(ctx, func(ctx context.Context, tx *sql.Tx) (done bool, err error) {
		db.logger.DebugContext(ctx, "Cleaning up alert logs...", "lastID", lastID)
		lastID, err = gadb.New(tx).CleanupAlertLogs(ctx,
			gadb.CleanupAlertLogsParams{
				BatchSize: batchSize,
				StartID:   lastID,
				EndID:     j.Args.EndID,
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
