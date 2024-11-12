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

		messages, err := q.SignalMgrGetPending(ctx)
		if err != nil {
			return fmt.Errorf("get pending signals: %w", err)
		}

		type dest struct {
			ServiceID uuid.UUID
			ChannelID uuid.UUID
		}

		res, err := q.SignalMgrGetScheduled(ctx)
		if err != nil {
			return fmt.Errorf("get scheduled signals: %w", err)
		}
		counts := make(map[dest]struct{}, len(res))
		for _, r := range res {
			counts[dest{ServiceID: r.ServiceID.UUID, ChannelID: r.ChannelID.UUID}] = struct{}{}
		}

		for _, m := range messages {
			if _, ok := counts[dest{ServiceID: m.ServiceID, ChannelID: m.DestID}]; ok {
				// Only allow one message per destination, per service, to be scheduled at a time.
				continue
			}
			id := uuid.New()
			err = q.SignalMgrInsertMessage(ctx, gadb.SignalMgrInsertMessageParams{
				ID:        id,
				ServiceID: uuid.NullUUID{Valid: true, UUID: m.ServiceID},
				ChannelID: uuid.NullUUID{Valid: true, UUID: m.DestID},
			})
			if err != nil {
				return fmt.Errorf("insert message: %w", err)
			}

			err = q.SignalMgrUpdateSignal(ctx, gadb.SignalMgrUpdateSignalParams{
				ID:        m.ID,
				MessageID: uuid.NullUUID{Valid: true, UUID: id},
			})
			if err != nil {
				return fmt.Errorf("update signal: %w", err)
			}
		}

		// We want to keep checking for work until we've processed everything,
		// even if we're waiting because there are already messages scheduled.
		hadWork = len(messages) > 0

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
