package alert

import (
	"time"
)

// State represents the current escalation state of an alert.
type State struct {
	// ID is the ID of the alert.
	ID             int
	StepNumber     int
	RepeatCount    int
	LastEscalation time.Time
}
