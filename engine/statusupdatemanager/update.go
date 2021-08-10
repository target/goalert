package statusupdatemanager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	alertlog "github.com/target/goalert/alert/log"
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
	// process up to 100
	for i := 0; i < 100; i++ {
		err = db.update(ctx)
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

func (db *DB) update(ctx context.Context) error {
	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer tx.Rollback()

	var id, alertID int
	var chanID, cmID sql.NullString
	var newStatus alert.Status
	err = tx.StmtContext(ctx, db.needsUpdate).QueryRowContext(ctx).Scan(&id, &chanID, &cmID, &alertID, &newStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return errDone
	}
	if err != nil {
		return fmt.Errorf("query out-of-date alert status: %w", err)
	}

	isSubscribed := chanID.Valid
	var userID sql.NullString
	if cmID.Valid {
		err = tx.StmtContext(ctx, db.cmWantsUpdates).QueryRowContext(ctx, cmID).Scan(&isSubscribed, &userID)
		if errors.Is(err, sql.ErrNoRows) {
			isSubscribed = false
			err = nil
		}
		if err != nil {
			return fmt.Errorf("check contact method status update config id='%s': %w", cmID.String, err)
		}
	}

	if isSubscribed {
		var logID int
		var event alertlog.Type
		switch newStatus {
		case alert.StatusTriggered:
			event = alertlog.TypeEscalated
		case alert.StatusActive:
			event = alertlog.TypeAcknowledged
		case alert.StatusClosed:
			event = alertlog.TypeClosed
		default:
			return fmt.Errorf("unknown alert status: %v", newStatus)
		}

		var logUserID sql.NullString
		err = tx.StmtContext(ctx, db.latestLogEntry).QueryRowContext(ctx, alertID, event).Scan(&logID, &logUserID)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			logID = 0
		}
		if err != nil {
			return fmt.Errorf("lookup latest log entry of '%s' for alert #%d: %w", event, alertID, err)
		}

		// Only insert message if the user is not the same as the log event user and we have a recent log entry.
		if logID > 0 && (!userID.Valid || userID.String != logUserID.String) {
			_, err = tx.StmtContext(ctx, db.insertMessage).ExecContext(ctx, uuid.New(), chanID, cmID, userID, alertID, logID)
			if err != nil {
				return fmt.Errorf("insert status update message for id=%d: %w", id, err)
			}
		}
	}

	if newStatus == alert.StatusClosed || !isSubscribed {
		_, err = tx.StmtContext(ctx, db.deleteSub).ExecContext(ctx, id)
		if err != nil {
			return fmt.Errorf("delete subscription for alert #%d (id=%d): %w", alertID, id, err)
		}
	} else {
		_, err = tx.StmtContext(ctx, db.updateStatus).ExecContext(ctx, id, newStatus)
		if err != nil {
			return fmt.Errorf("update status for alert #%d to '%s' (id=%d): %w", alertID, newStatus, id, err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
