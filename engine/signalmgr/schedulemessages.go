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

func (db *DB) scheduleMessages(ctx context.Context, serviceID uuid.NullUUID) error {
	var didWork bool
	err := db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
		q := gadb.NewCompat(tx)

		messages, err := q.SignalMgrGetPending(ctx, serviceID)
		if err != nil {
			return fmt.Errorf("get pending signals: %w", err)
		}

		type dest struct {
			ServiceID uuid.UUID
			ChannelID uuid.UUID
		}

		res, err := q.SignalMgrGetScheduled(ctx, serviceID)
		if err != nil {
			return fmt.Errorf("get scheduled signals: %w", err)
		}
		alreadyScheduled := make(map[dest]struct{}, len(res))
		for _, r := range res {
			alreadyScheduled[dest{ServiceID: r.ServiceID.UUID, ChannelID: r.ChannelID.UUID}] = struct{}{}
		}

		for _, m := range messages {
			if _, ok := alreadyScheduled[dest{ServiceID: m.ServiceID, ChannelID: m.DestID}]; ok {
				// Only allow one message per destination, per service, to be scheduled at a time.
				continue
			}
			didWork = true
			alreadyScheduled[dest{ServiceID: m.ServiceID, ChannelID: m.DestID}] = struct{}{}
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

		return nil
	})
	if err != nil {
		return err
	}

	if serviceID.Valid {
		// only try once for per-service/on-demand updates
		return nil
	}

	if didWork {
		// for global update, we want to keep checking for work until nothing is left.
		return river.JobSnooze(time.Second)
	}

	return nil
}
