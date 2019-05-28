package notification

// Alert represents outgoing notifications for alerts.
type Alert struct {
	Dest       Dest
	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
	AlertID    int    // The global alert number
	Summary    string
	Details    string
}

var _ Message = &Alert{}

func (a Alert) Type() MessageType    { return MessageTypeAlert }
func (a Alert) ID() string           { return a.CallbackID }
func (a Alert) Destination() Dest    { return a.Dest }
func (a Alert) Body() string         { return a.Summary }
func (a Alert) ExtendedBody() string { return a.Details }
func (a Alert) SubjectID() int       { return a.AlertID }
