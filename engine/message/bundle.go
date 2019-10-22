package message

import (
	"context"
	"database/sql"
)

func (db *DB) bundleMessages(ctx context.Context, tx *sql.Tx, msg []Message) ([]Message, error) {
	// TODO
	return msg, nil
}
