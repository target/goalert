package notification

import (
	"time"
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

	UserID          string // might need to join user details
	ContactMethodID string // might need to join CM details
	ChannelID       string // might need to join channel details
	ServiceID       string // might need to join service details
}
