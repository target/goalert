package signal

import (
	"time"

	"github.com/target/goalert/notification"
)

// Signal represents the data for an outgoing signal.
type OutgoingSignal struct {
	ID   string
	Type notification.MessageType
	Dest notification.Dest

	SignalID  int
	UserID    string
	ServiceID string
	Message   string
	CreatedAt time.Time
	SentAt    time.Time
}
