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

	_, err := db.lock.Exec(ctx, db.insertMessages)
	if err != nil {
		return errors.Wrap(err, "insert status update messages")
	}

	return nil
}
