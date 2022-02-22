package notification

import "github.com/target/goalert/notification/nfynet"

// AlertState is the current state of an Alert.
type AlertState int

// All alert states
const (
	AlertStateUnknown AlertState = iota
	AlertStateUnacknowledged
	AlertStateAcknowledged
	AlertStateClosed
)

type AlertStatus struct {
	nfynet.Target

	CallbackID string
	AlertID    int
	LogEntry   string

	// Summary of the alert that this status is in regards to.
	Summary string
	// Details of the alert that this status is in regards to.
	Details string

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus SendResult

	// NewAlertState contains the most recent state of the alert.
	NewAlertState AlertState
}

var _ Message = &AlertStatus{}

func (s AlertStatus) ID() string { return s.CallbackID }
