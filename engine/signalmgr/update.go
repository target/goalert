package signalmgr

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
)

var _ processinglock.Updatable = &DB{}

func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	return db.lock.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		q := gadb.New(tx)

		err := q.SignalMgrDeleteStale(ctx)
		if err != nil {
			return fmt.Errorf("delete stale signals: %w", err)
		}

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

		return nil
	})
}
