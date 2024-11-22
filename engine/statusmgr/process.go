package statusmgr

import (
	"context"
	"database/sql"

	"github.com/riverqueue/river"
)

// ProcessArgs is the arguments for processing a single subscription.
type ProcessArgs struct {
	SubscriptionID int64
}

func (ProcessArgs) Kind() string { return "status-manager-process-subscription" }
func (db *DB) processSubscription(ctx context.Context, j *river.Job[ProcessArgs]) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		ctx = db.cfgSrc.Config().Context(ctx) // mix in current config

		return db.update(ctx, tx, j.Args.SubscriptionID)
	})
}
