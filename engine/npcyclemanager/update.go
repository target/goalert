package npcyclemanager

import (
	"context"
	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

// UpdateAll will update and cleanup all notification cycles.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := db.update(ctx, true, nil)
	return err
}

// UpdateOneAlert will update and cleanup all notification cycles for the given alert.
func (db *DB) UpdateOneAlert(ctx context.Context, alertID int) error {
	ctx = log.WithField(ctx, "AlertID", alertID)
	return db.update(ctx, false, &alertID)
}

func (db *DB) update(ctx context.Context, all bool, alertID *int) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Updating notification cycles.")


	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.StmtContext(ctx, db.queueMessages).QueryContext(ctx)
	if err != nil {
		return err
	}
	defer rows.Close()

	type record struct {
		alertID int
		userID string
		// TODO: populate meta data
		meta *alertlog.NotificationMetaData
	}

	var data []record
	for rows.Next() {
		var rec record
		err = rows.Scan(&rec.userID, &rec.alertID)
		if err != nil {
			return err
		}
		data = append(data, rec)
	}

	for _, rec := range data {
		err = db.log.LogTx(ctx, tx, rec.alertID, alertlog.TypeNoNotificationSent, rec.meta)
		if err != nil {
			return errors.Wrap(err, "queue outgoing messages")
		}
	}

	return tx.Commit()
}
