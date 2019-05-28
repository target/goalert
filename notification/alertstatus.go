package notification

type AlertStatus struct {
	Dest      Dest
	MessageID string
	AlertID   int
	Log       string
}

var _ Message = &AlertStatus{}

func (s AlertStatus) Type() MessageType    { return MessageTypeAlertStatus }
func (s AlertStatus) ID() string           { return s.MessageID }
func (s AlertStatus) Destination() Dest    { return s.Dest }
func (s AlertStatus) Body() string         { return s.Log }
func (s AlertStatus) ExtendedBody() string { return "" }
func (s AlertStatus) SubjectID() int       { return s.AlertID }
