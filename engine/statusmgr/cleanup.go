package statusmgr

import (
	"context"
	"database/sql"

	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

type CleanupArgs struct{}

func (CleanupArgs) Kind() string { return "status-manager-cleanup" }

func (db *DB) cleanup(ctx context.Context, j *river.Job[CleanupArgs]) error {
	return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		return gadb.New(tx).StatusMgrCleanupStaleSubs(ctx)
	})
}
