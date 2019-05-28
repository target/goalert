package notification

// Verification represents outgoing verification code.
type Verification struct {
	Dest       Dest
	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
	Code       int
}

var _ Message = &Test{}

func (v Verification) Type() MessageType    { return MessageTypeVerification }
func (v Verification) ID() string           { return v.CallbackID }
func (v Verification) Destination() Dest    { return v.Dest }
func (v Verification) Body() string         { return "" }
func (v Verification) ExtendedBody() string { return "" }
func (v Verification) SubjectID() int       { return v.Code }
