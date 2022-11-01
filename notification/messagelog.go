package notification

import (
	"time"

	"github.com/google/uuid"
)

type MessageLog struct {
	ID           string
	CreatedAt    time.Time
	LastStatusAt time.Time
	MessageType  MessageType

	LastStatus    State
	StatusDetails string
	SrcValue      string

	AlertID       int
	ProviderMsgID *ProviderMessageID

	UserID   string
	UserName string

	ContactMethodID string

	ChannelID uuid.UUID

	ServiceID   string
	ServiceName string
}
