package statusmgr

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util/log"
)

func (db *DB) update(ctx context.Context, tx *sql.Tx, id int64) error {
	q := gadb.New(tx)

	sub, err := q.StatusMgrFindOne(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		// subscription was deleted or locked by another job
		return nil
	}
	if err != nil {
		return fmt.Errorf("query out-of-date alert status: %w", err)
	}

	if sub.LastAlertStatus == sub.Status {
		// fast path, no update needed (possible duplicate job)
		return nil
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

	entry, err := q.StatusMgrLogEntry(ctx, gadb.StatusMgrLogEntryParams{
		AlertID:   sub.AlertID,
		EventType: eventType,
	})
	if errors.Is(err, sql.ErrNoRows) {
		// no log entry, ignore
		err = nil
	}
	if err != nil {
		return fmt.Errorf("lookup latest log entry of '%s' for alert #%d: %w", eventType, sub.AlertID, err)
	}

	switch {
	case entry.ID == 0:
		// no log entry, log error but continue
		log.Log(ctx, fmt.Errorf("no log entry found for alert #%d status update (%s), skipping", sub.AlertID, eventType))
	case sub.ContactMethodID.Valid:
		info, err := q.ContactMethodFineOne(ctx, sub.ContactMethodID.UUID)
		if errors.Is(err, sql.ErrNoRows) || info.Disabled {
			// contact method was deleted or disabled
			return q.StatusMgrDeleteSub(ctx, sub.ID)
		}
		if err != nil {
			return fmt.Errorf("lookup contact method info: %w", err)
		}
		typeInfo, err := db.reg.TypeInfo(ctx, info.Dest.DestV1.Type)
		if err != nil {
			return fmt.Errorf("lookup contact method type info: %w", err)
		}
		if !typeInfo.SupportsStatusUpdates {
			// contact method doesn't support status updates
			return q.StatusMgrDeleteSub(ctx, sub.ID)
		}
		forceUpdate := typeInfo.SupportsStatusUpdates && typeInfo.StatusUpdatesRequired
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
			LogID:   sql.NullInt64{Int64: int64(entry.ID), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("send user status update message: %w", err)
		}
	case sub.ChannelID.Valid:
		err = q.StatusMgrSendChannelMsg(ctx, gadb.StatusMgrSendChannelMsgParams{
			ID:        uuid.New(),
			ChannelID: sub.ChannelID.UUID,
			AlertID:   sub.AlertID,
			LogID:     sql.NullInt64{Int64: int64(entry.ID), Valid: true},
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
