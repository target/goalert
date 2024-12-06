package statusmgr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

type LookForWorkArgs struct {
	AlertID int64 `json:",omitempty"`
}

func (LookForWorkArgs) Kind() string { return "status-manager-look-for-work" }

// lookForWork is a worker function that will find any subscriptions that are out of date and need to be updated, and add them to the processing queue.
func (db *DB) lookForWork(ctx context.Context, j *river.Job[LookForWorkArgs]) error {
	var outOfDate []int64
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		outOfDate, err = gadb.New(tx).StatusMgrOutdated(ctx, sql.NullInt64{Int64: j.Args.AlertID, Valid: j.Args.AlertID != 0})
		return err
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(outOfDate) == 0 {
		return nil
	}

	params := make([]river.InsertManyParams, len(outOfDate))
	for i, id := range outOfDate {
		params[i] = river.InsertManyParams{
			Args: ProcessArgs{SubscriptionID: id},
			InsertOpts: &river.InsertOpts{
				Queue:    QueueName,
				Priority: PriorityProcess,
			},
		}
	}

	r := river.ClientFromContext[pgx.Tx](ctx)
	_, err = r.InsertManyFast(ctx, params)
	if err != nil {
		return fmt.Errorf("insert many: %w", err)
	}

	return nil
}
