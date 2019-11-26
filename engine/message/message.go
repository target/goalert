package message

import (
	"github.com/target/goalert/notification"
)

// Message represents the data for an outgoing message.
type Message struct {
	ID         string
	Type       Type
	DestType   notification.DestType
	DestID     string
	AlertID    int
	AlertLogID int
	VerifyID   string
}
