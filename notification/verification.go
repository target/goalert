package notification

import "github.com/target/goalert/gadb"

// Verification represents outgoing verification code.
type Verification struct {
	Dest       gadb.DestV1
	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
	Code       int
}

var _ Message = &Verification{}

func (v Verification) Type() MessageType    { return MessageTypeVerification }
func (v Verification) ID() string           { return v.CallbackID }
func (v Verification) Body() string         { return "" }
func (v Verification) ExtendedBody() string { return "" }
func (v Verification) SubjectID() int       { return v.Code }

func (v Verification) Destination() gadb.DestV1 { return v.Dest }
