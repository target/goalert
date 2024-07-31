package notification

import "github.com/target/goalert/gadb"

// Alert represents outgoing notifications for alerts.
type Alert struct {
	Dest        gadb.DestV1
	CallbackID  string // CallbackID is the identifier used to communicate a response to the notification
	AlertID     int    // The global alert number
	Summary     string
	Details     string
	ServiceID   string
	ServiceName string
	Meta        map[string]string

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus *SendResult
}

type AlertPendingNotification struct {
	DestName string
	DestType string
}

var _ Message = &Alert{}

func (a Alert) Type() MessageType    { return MessageTypeAlert }
func (a Alert) ID() string           { return a.CallbackID }
func (a Alert) Body() string         { return a.Summary }
func (a Alert) ExtendedBody() string { return a.Details }
func (a Alert) SubjectID() int       { return a.AlertID }

func (a Alert) Destination() gadb.DestV1 { return a.Dest }
