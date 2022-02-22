package notification

import "github.com/target/goalert/notification/nfynet"

// Verification represents outgoing verification code.
type Verification struct {
	nfynet.Target

	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
	Code       int
}

var _ Message = &Verification{}

func (v Verification) ID() string { return v.CallbackID }
