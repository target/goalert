package notification

import (
	"github.com/target/goalert/notification/nfy"
)

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	Type() MessageType

	DestType() nfy.DestType
	DestArg(string) string
	DestHash() nfy.DestHash
}
