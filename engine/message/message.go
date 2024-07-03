package message

import (
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/notification"
)

// Message represents the data for an outgoing message.
type Message struct {
	ID         string
	Type       notification.MessageType
	Dest       notification.Dest
	CMID       uuid.NullUUID
	NCID       uuid.NullUUID
	AlertID    int
	AlertLogID int
	VerifyID   string

	UserID     string
	ServiceID  string
	ScheduleID string
	CreatedAt  time.Time
	SentAt     time.Time

	StatusAlertIDs []int64
}
