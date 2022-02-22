package notification

import "github.com/target/goalert/notification/nfynet"

// Alert represents outgoing notifications for alerts.
type Alert struct {
	nfynet.Target

	CallbackID string // CallbackID is the identifier used to communicate a response to the notification
	AlertID    int    // The global alert number
	Summary    string
	Details    string

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus *SendResult
}

type AlertPendingNotification struct {
	DestName string
	DestType string
}

var _ Message = &Alert{}

func (a Alert) ID() string { return a.CallbackID }
