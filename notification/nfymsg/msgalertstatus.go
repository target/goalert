package nfymsg

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
	Base

	AlertID   int
	LogEntry  string
	ServiceID string

	// Summary of the alert that this status is in regards to.
	Summary string
	// Details of the alert that this status is in regards to.
	Details string

	// OriginalStatus is the status of the first Alert notification to this Dest for this AlertID.
	OriginalStatus SendResult

	// NewAlertState contains the most recent state of the alert.
	NewAlertState AlertState
}
