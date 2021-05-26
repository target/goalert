package message

import (
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/notification"
)

// Message represents the data for an outgoing message.
type Message struct {
	ID         string
	Type       notification.MessageType
	Dest       notification.Dest
	AlertID    int
	AlertLogID int
	VerifyID   string

	UserID     string
	ServiceID  string
	ScheduleID uuid.UUID
	CreatedAt  time.Time
	SentAt     time.Time

	StatusAlertIDs []int
}
