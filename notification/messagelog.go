package notification

import (
	"time"

	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/service"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
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

	User          user.User
	ContactMethod contactmethod.ContactMethod
	Channel       notificationchannel.Channel
	Service       service.Service
}
