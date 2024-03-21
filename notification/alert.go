package notification

// Alert represents outgoing notifications for alerts.
type Alert struct {
	Dest        Dest
	CallbackID  string // CallbackID is the identifier used to communicate a response to the notification
	AlertID     int    // The global alert number
	Summary     string
	Details     string
	ServiceID   string
	ServiceName string
	Meta        []AlertMeta

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus *SendResult
}

type AlertMeta struct {
	Key   string
	Value string
}

type AlertPendingNotification struct {
	DestName string
	DestType string
}

var _ Message = &Alert{}

func (a Alert) Type() MessageType    { return MessageTypeAlert }
func (a Alert) ID() string           { return a.CallbackID }
func (a Alert) Destination() Dest    { return a.Dest }
func (a Alert) Body() string         { return a.Summary }
func (a Alert) ExtendedBody() string { return a.Details }
func (a Alert) SubjectID() int       { return a.AlertID }
