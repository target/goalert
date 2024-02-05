package statusmgr

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/log"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "Processing status updates.")

	err = db.lock.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		q := gadb.New(tx)

		err := q.StatusMgrUpdateCMForced(ctx)
		if err != nil {
			return fmt.Errorf("update contact methods for forced status updates: %w", err)
		}

		err = q.StatusMgrCleanupDisabledSubs(ctx)
		if err != nil {
			return fmt.Errorf("delete status subscriptions for disabled contact methods: %w", err)
		}

		err = q.StatusMgrCleanupStaleSubs(ctx)
		if err != nil {
			return fmt.Errorf("delete stale status subscriptions: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// process up to 100
	for i := 0; i < 100; i++ {
		err = db.lock.WithTx(ctx, db.update)
		if errors.Is(err, errDone) {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

var errDone = errors.New("done")

func (db *DB) update(ctx context.Context, tx *sql.Tx) error {
	q := gadb.New(tx)

	sub, err := q.StatusMgrNextUpdate(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return errDone
	}
	if err != nil {
		return fmt.Errorf("query out-of-date alert status: %w", err)
	}

	var eventType gadb.EnumAlertLogEvent
	switch sub.Status {
	case gadb.EnumAlertStatusTriggered:
		eventType = gadb.EnumAlertLogEventEscalated
	case gadb.EnumAlertStatusActive:
		eventType = gadb.EnumAlertLogEventAcknowledged
	case gadb.EnumAlertStatusClosed:
		eventType = gadb.EnumAlertLogEventClosed
	}

	var entryID sql.NullInt64
	entry, err := q.StatusMgrLogEntry(ctx, gadb.StatusMgrLogEntryParams{
		AlertID:   sub.AlertID,
		EventType: eventType,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// log entry is best-effort, but not required
		err = nil
	}
	if err != nil {
		return fmt.Errorf("lookup latest log entry of '%s' for alert #%d: %w", eventType, sub.AlertID, err)
	}
	if entry.ID > 0 {
		entryID = sql.NullInt64{Int64: int64(entry.ID), Valid: true}
	}

	switch {
	case sub.ContactMethodID.Valid:
		info, err := q.StatusMgrCMInfo(ctx, sub.ContactMethodID.UUID)
		if errors.Is(err, sql.ErrNoRows) {
			// contact method was deleted or disabled
			return q.StatusMgrDeleteSub(ctx, sub.ID)
		}
		if err != nil {
			return fmt.Errorf("lookup contact method info: %w", err)
		}
		forceUpdate := contactmethod.Type(info.Type).StatusUpdatesAlways()
		if !forceUpdate && entry.UserID.UUID == info.UserID {
			// We don't want to update a user for their own actions, unless
			// the contact method is forced to always send updates.
			break
		}
		err = q.StatusMgrSendUserMsg(ctx, gadb.StatusMgrSendUserMsgParams{
			ID:      uuid.New(),
			CmID:    sub.ContactMethodID.UUID,
			AlertID: sub.AlertID,
			UserID:  info.UserID,
			LogID:   entryID,
		})
		if err != nil {
			return fmt.Errorf("send user status update message: %w", err)
		}
	case sub.ChannelID.Valid:
		err = q.StatusMgrSendChannelMsg(ctx, gadb.StatusMgrSendChannelMsgParams{
			ID:        uuid.New(),
			ChannelID: sub.ChannelID.UUID,
			AlertID:   sub.AlertID,
			LogID:     entryID,
		})
		if err != nil {
			return fmt.Errorf("send channel status update message: %w", err)
		}
	default:
		return fmt.Errorf("invalid subscription: %v", sub)
	}

	if sub.Status == gadb.EnumAlertStatusClosed {
		return q.StatusMgrDeleteSub(ctx, sub.ID)
	}

	return q.StatusMgrUpdateSub(ctx, gadb.StatusMgrUpdateSubParams{
		ID:              sub.ID,
		LastAlertStatus: sub.Status,
	})
}
