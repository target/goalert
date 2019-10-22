package message

import (
	"context"
	"database/sql"
)

// bundleAlertMessages will bundle status updates for the same Dest value. It will not attempt to join existing status-update bundles.
//
// It also handles updating the outgoing_messages table by marking bundled messages with the `bundled`
// status and creating a new bundled message placeholder.
func (db *DB) bundleAlertMessages(ctx context.Context, tx *sql.Tx, messages []Message) ([]Message, error) {
	return messages, nil
}
