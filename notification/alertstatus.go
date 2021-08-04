package notification

type AlertStatus struct {
	Dest       Dest
	CallbackID string
	AlertID    int
	LogEntry   string

	// Summary of the alert that this status is in regards to.
	Summary string
	// Details of the alert that this status is in regards to.
	Details string

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus SendResult

	// Status is the most recent status of the Alert.
	Status string
}

var _ Message = &AlertStatus{}

func (s AlertStatus) Type() MessageType { return MessageTypeAlertStatus }
func (s AlertStatus) ID() string        { return s.CallbackID }
func (s AlertStatus) Destination() Dest { return s.Dest }
