package statusmgr

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	log.Debugf(ctx, "Processing status updates.")

	types, err := db.reg.Types(ctx)
	if err != nil {
		return fmt.Errorf("get notification destination types: %w", err)
	}

	var forcedTypes []string
	for _, t := range types {
		if t.SupportsStatusUpdates && t.StatusUpdatesRequired {
			forcedTypes = append(forcedTypes, t.Type)
		}
	}

	err = db.lock.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		q := gadb.NewCompat(tx)

		err := q.StatusMgrUpdateCMForced(ctx, forcedTypes)
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

	// Clear omit list, as we want to process all
	// subscriptions in the next step.
	//
	// We don't want to assign nil to omit, as it
	// will cause the query to fail.
	db.omit = db.omit[:0]

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
	q := gadb.NewCompat(tx)

	sub, err := q.StatusMgrNextUpdate(ctx, db.omit)
	if errors.Is(err, sql.ErrNoRows) {
		return errDone
	}
	if err != nil {
		return fmt.Errorf("query out-of-date alert status: %w", err)
	}

	// Add to omit list to prevent re-processing
	// the same subscription in the same run.
	db.omit = append(db.omit, sub.ID)

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
		info, err := q.StatusMgrCMInfo(ctx, sub.ContactMethodID.UUID)
		if errors.Is(err, sql.ErrNoRows) {
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
			LogID:   pgtype.Int8{Int64: int64(entry.ID), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("send user status update message: %w", err)
		}
	case sub.ChannelID.Valid:
		err = q.StatusMgrSendChannelMsg(ctx, gadb.StatusMgrSendChannelMsgParams{
			ID:        uuid.New(),
			ChannelID: sub.ChannelID.UUID,
			AlertID:   sub.AlertID,
			LogID:     pgtype.Int8{Int64: int64(entry.ID), Valid: true},
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
