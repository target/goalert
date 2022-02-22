package notification

import "github.com/target/goalert/notification/nfynet"

// Test represents outgoing test notification.
type Test struct {
	nfynet.Target

	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
}

var _ Message = &Test{}

func (t Test) ID() string { return t.CallbackID }
