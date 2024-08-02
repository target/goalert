package nfydest

import (
	"context"

	"github.com/target/goalert/notification/nfymsg"
)

// A MessageStatuser is an optional interface a Sender can implement that allows checking the status
// of a previously sent message by it's externalID.
type MessageStatuser interface {
	MessageStatus(ctx context.Context, externalID string) (*nfymsg.Status, error)
}
