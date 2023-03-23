package statusupdatemanager

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
		err := gadb.New(tx).StatusMgrForcedStatusUpdates(ctx)
		if err != nil {
			return fmt.Errorf("update status update enable: %w", err)
		}
		err = gadb.New(tx).StatusMgrCleanupSubscriptions(ctx)
		if err != nil {
			return fmt.Errorf("delete status subscriptions for disabled contact methods: %w", err)
		}
		return nil
	})
	if err != nil {
		// okay to proceed
		log.Log(ctx, err)
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

	sub, err := q.StatusMgrNextSubscription(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return errDone
	}
	if err != nil {
		return fmt.Errorf("query next status subscription: %w", err)
	}

	var logEvent gadb.EnumAlertLogEvent
	switch sub.Status {
	case gadb.EnumAlertStatusTriggered:
		logEvent = gadb.EnumAlertLogEventEscalated
	case gadb.EnumAlertStatusActive:
		logEvent = gadb.EnumAlertLogEventAcknowledged
	case gadb.EnumAlertStatusClosed:
		logEvent = gadb.EnumAlertLogEventClosed
	}

	entry, err := q.StatusMgrLastLog(ctx, gadb.StatusMgrLastLogParams{
		AlertID:   sub.AlertID,
		EventType: logEvent,
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("query last log entry: %w", err)
	}
	var logID sql.NullInt64
	if entry.ID > 0 {
		logID = sql.NullInt64{Int64: int64(entry.ID), Valid: true}
	}

	switch {
	case sub.ContactMethodID.Valid:
		info, err := q.StatusMgrCMInfo(ctx, sub.ContactMethodID.UUID)
		if errors.Is(err, sql.ErrNoRows) {
			// Delete it since the contact method was deleted, disabled, or has updates turned off.
			return q.StatusMgrDelete(ctx, sub.ID)
		}
		forceUpdate := contactmethod.Type(info.Type).StatusUpdatesAlways()
		if forceUpdate || !entry.SubUserID.Valid || entry.SubUserID.UUID != info.UserID {
			err = q.StatusMgrInsertUserCMMessage(ctx, gadb.StatusMgrInsertUserCMMessageParams{
				MsgID:      uuid.New(),
				AlertID:    sub.AlertID,
				CmID:       sub.ContactMethodID.UUID,
				UserID:     info.UserID,
				AlertLogID: logID,
			})
			if err != nil {
				return fmt.Errorf("insert user contact method message: %w", err)
			}
		}
	case sub.ChannelID.Valid:
		err = q.StatusMgrInsertChanMessage(ctx, gadb.StatusMgrInsertChanMessageParams{
			MsgID:      uuid.New(),
			AlertID:    sub.AlertID,
			ChannelID:  sub.ChannelID.UUID,
			AlertLogID: logID,
		})
		if err != nil {
			return fmt.Errorf("insert channel message: %w", err)
		}
	default:
		return fmt.Errorf("invalid subscription: %v", sub)
	}

	if sub.Status == gadb.EnumAlertStatusClosed {
		// Since the alert is closed, there can be no more updates.
		return q.StatusMgrDelete(ctx, sub.ID)
	}

	return q.StatusMgrUpdate(ctx, gadb.StatusMgrUpdateParams{
		ID:              sub.ID,
		LastAlertStatus: sub.Status,
	})
}
