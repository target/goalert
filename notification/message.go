package notification

import "github.com/target/goalert/gadb"

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	Type() MessageType
	Destination() gadb.DestV1
}
