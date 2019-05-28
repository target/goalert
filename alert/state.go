package alert

import (
	"time"
)

// State represents the current escalation state of an alert.
type State struct {
	AlertID        int
	StepNumber     int
	RepeatCount    int
	LastEscalation time.Time
}
