package notification

type AlertStatusBundle struct {
	Dest         Dest
	MessageID    string
	Log          string
	AlertID      int
	OtherUpdates int
}

var _ Message = &AlertStatusBundle{}

func (b AlertStatusBundle) Type() MessageType { return MessageTypeAlertStatusBundle }
func (b AlertStatusBundle) ID() string        { return b.MessageID }
func (b AlertStatusBundle) Destination() Dest { return b.Dest }
