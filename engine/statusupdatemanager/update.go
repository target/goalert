package statusupdatemanager

import (
	"context"

	"github.com/pkg/errors"
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
	_, err = db.lock.Exec(ctx, db.insertUserMessages)
	if err != nil {
		return errors.Wrap(err, "insert user status update messages")
	}
	_, err = db.lock.Exec(ctx, db.insertChannelMessages)
	if err != nil {
		return errors.Wrap(err, "insert notification channel status update messages")
	}

	return nil
}
