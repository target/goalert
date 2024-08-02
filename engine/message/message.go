package message

import (
	"time"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
)

// Message represents the data for an outgoing message.
type Message struct {
	ID         string
	Type       gadb.EnumOutgoingMessagesType
	DestID     notification.DestID
	Dest       gadb.DestV1
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

func (m Message) Base() nfymsg.Base {
	return nfymsg.Base{
		ID:   m.ID,
		Dest: m.Dest,
	}
}
