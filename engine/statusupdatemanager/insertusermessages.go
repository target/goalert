package statusupdatemanager

import (
	"context"

	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

func (db *DB) updateUsers(ctx context.Context, all bool, alertID *int) error {
	log.Debugf(ctx, "Processing status updates.")

	_, err := db.lock.Exec(ctx, db.insertUserMessages)
	if err != nil {
		return errors.Wrap(err, "insert status update messages")
	}

	return nil
}
