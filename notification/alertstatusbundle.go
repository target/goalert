package notification

type AlertStatusBundle struct {
	Dest       Dest
	CallbackID string
	LogEntry   string
	AlertID    int
	Count      int // The total number of status updates this bundle represents.
}

var _ Message = &AlertStatusBundle{}

func (b AlertStatusBundle) Type() MessageType { return MessageTypeAlertStatusBundle }
func (b AlertStatusBundle) ID() string        { return b.CallbackID }
func (b AlertStatusBundle) Destination() Dest { return b.Dest }
