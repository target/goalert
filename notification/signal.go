package notification

// Signal represents outgoing notifications for Signals
type Signal struct {
	Dest        Dest
	CallbackID  string // CallbackID is the identifier used to communicate a response to the notification
	SignalID    int    // The global signal number
	Summary     string
	ServiceID   string
	ServiceName string

	// Email is an optional field containing additional info required for sending signals via email
	Email *SignalEmail

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus *SendResult
}

type SignalEmail struct {
	Subject string
	Body    string
}

var _ Message = &Signal{}

func (n Signal) ID() string        { return n.CallbackID }
func (n Signal) Destination() Dest { return n.Dest }
func (n Signal) Type() MessageType { return MessageTypeSignal }
