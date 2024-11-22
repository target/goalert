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

type LookForWorkArgs struct{}

func (LookForWorkArgs) Kind() string { return "status-manager-look-for-work" }

func (db *DB) lookForWork(ctx context.Context, j *river.Job[LookForWorkArgs]) error {
	var outOfDate []int64
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		outOfDate, err = gadb.New(tx).StatusMgrOutdated(ctx)
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
