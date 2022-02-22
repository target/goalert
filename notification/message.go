package notification

import "github.com/target/goalert/notification/nfynet"

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	TargetID() nfynet.TargetID
}
