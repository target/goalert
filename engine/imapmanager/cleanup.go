package imapmanager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util"
)

// CleanupArgs are the arguments for the cleanup job.
type CleanupArgs struct{}

func (CleanupArgs) Kind() string { return "imap-cleanup" }

// CleanupProcessedMessages removes old processed message records.
func (db *DB) CleanupProcessedMessages(ctx context.Context, job *river.Job[CleanupArgs]) error {
	db.logger.Info("IMAP cleanup: starting")

	// Run cleanup in batches until no more work
	var totalDeleted int64
	for {
		var deleted int64
		err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
			queries := gadb.New(tx)
			rows, err := queries.IMAPCleanupProcessedMessages(ctx)
			deleted = rows
			return err
		})

		if err != nil {
			return fmt.Errorf("cleanup processed messages: %w", err)
		}

		totalDeleted += deleted

		if deleted == 0 {
			// No more work
			break
		}

		// Sleep briefly between batches to avoid overwhelming the database
		err = util.ContextSleep(ctx, 100)
		if err != nil {
			return err
		}
	}

	db.logger.Info("IMAP cleanup: completed", "deleted", totalDeleted)
	return nil
}
