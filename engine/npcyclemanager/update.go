package npcyclemanager

import (
	"context"
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

	_, err = db.lock.Exec(ctx, db.queueMessages)
	return errors.Wrap(err, "queue outgoing messages")
}
