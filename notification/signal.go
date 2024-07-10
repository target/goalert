package notification

// SignalMessage is a dynamic message that is sent to a notification destination.
type SignalMessage struct {
	Dest       Dest
	CallbackID string // CallbackID is the identifier used to communicate a response to the notification

	Params map[string]string
}

var _ Message = &Test{}

func (t SignalMessage) Type() MessageType { return MessageTypeSignalMessage }
func (t SignalMessage) ID() string        { return t.CallbackID }
func (t SignalMessage) Destination() Dest { return t.Dest }
func (t SignalMessage) Param(name string) string {
	if t.Params == nil {
		return ""
	}
	return t.Params[name]
}
func (SignalMessage) Body() string         { return "" }
func (SignalMessage) ExtendedBody() string { return "" }
func (SignalMessage) SubjectID() int       { return -1 }
