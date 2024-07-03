package notification

// AlertBundle represents a bundle of outgoing alert notifications for a single service.
type AlertBundle struct {
	DestV2
	CallbackID  string // CallbackID is the identifier used to communicate a response to the notification
	ServiceID   string
	ServiceName string // The service being notified for
	Count       int    // Number of unacked alerts
}

var _ Message = &AlertBundle{}

func (b AlertBundle) Type() MessageType { return MessageTypeAlertBundle }
func (b AlertBundle) ID() string        { return b.CallbackID }
