package verifymanager

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"

	"github.com/pkg/errors"
)

// UpdateAll will insert all verification requests into outgoing_messages.
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	err = db.update(ctx)
	return err
}
func (db *DB) update(ctx context.Context) error {
	log.Debugf(ctx, "Processing verification messages.")
	var err error
	exec := func(s *sql.Stmt, msg string) bool {
		_, err = db.lock.Exec(ctx, s)
		if err != nil {
			err = errors.Wrap(err, msg)
			return false
		}
		return true
	}

	if !exec(db.cleanupExpired, "cleanup expired codes") {
		return err
	}

	if !exec(db.insertMessages, "insert messages") {
		return err
	}

	return nil
}
