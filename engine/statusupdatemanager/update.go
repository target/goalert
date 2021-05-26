package statusupdatemanager

import (
	"context"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

// UpdateAll will update all schedule rules.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = db.update(ctx, true, nil)
	return err
}

func (db *DB) update(ctx context.Context, all bool, alertID *int) error {
	log.Debugf(ctx, "Processing status updates.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "start transaction")
	}
	defer tx.Rollback()
	log.Debugf(ctx, "Updating outgoing messages.")

	rows, err := tx.StmtContext(ctx, db.needsUpdate).QueryContext(ctx)
	if err != nil {
		return errors.Wrap(err, "needs update messages")
	}
	defer rows.Close()

	for rows.Next() {
		var id, channel_id, contact_method_id, status string
		var alert_id int
		err = rows.Scan(&id, &channel_id, &contact_method_id, &alert_id, &status)
		if err != nil {
			return errors.Wrap(err, "scan alerts subscriptions data")
		}

		_, err := tx.StmtContext(ctx, db.insertMessages).ExecContext(ctx, id, channel_id, contact_method_id, alert_id, status)
		if err != nil {
			return errors.Wrap(err, "insert outgoing messages")
		}

		_, err = tx.StmtContext(ctx, db.updateStatus).ExecContext(ctx, status, alert_id)
		if err != nil {
			return errors.Wrap(err, "update status")
		}

		_, err = tx.StmtContext(ctx, db.cleanupClosed).ExecContext(ctx, alert_id)
		if err != nil {
			return errors.Wrap(err, "update status")
		}
	}

	return nil
}
