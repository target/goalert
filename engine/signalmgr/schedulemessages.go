package signalmgr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/target/goalert/gadb"
)

func (db *DB) scheduleMessages(ctx context.Context) error {
	var hadWork bool
	err := db.lock.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		q := gadb.New(tx)

		work, err := q.SignalMgrGetPending(ctx)
		if err != nil {
			return fmt.Errorf("get pending signals: %w", err)
		}

		for _, w := range work {
			id := uuid.New()
			err = q.SignalMgrInsertMessage(ctx, gadb.SignalMgrInsertMessageParams{
				ID:        id,
				ServiceID: uuid.NullUUID{Valid: true, UUID: w.ServiceID},
				ChannelID: uuid.NullUUID{Valid: true, UUID: w.DestID},
			})
			if err != nil {
				return fmt.Errorf("insert message: %w", err)
			}

			err = q.SignalMgrUpdateSignal(ctx, gadb.SignalMgrUpdateSignalParams{
				ID:        w.ID,
				MessageID: uuid.NullUUID{Valid: true, UUID: id},
			})
			if err != nil {
				return fmt.Errorf("update signal: %w", err)
			}
		}
		hadWork = len(work) > 0

		return nil
	})
	if err != nil {
		return err
	}

	if hadWork {
		// reschedule to finish processing
		return river.JobSnooze(time.Second)
	}

	return nil
}
