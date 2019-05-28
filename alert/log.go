package alert

import (
	"fmt"
	"time"
)

// A LogEvent represents a state change of an alert.
type LogEvent string

// Types of LogEvents
const (
	LogEventCreated           LogEvent = "created"
	LogEventReopened          LogEvent = "reopened"
	LogEventClosed            LogEvent = "closed"
	LogEventStatusChanged     LogEvent = "status_changed"
	LogEventAssignmentChanged LogEvent = "assignment_changed"
	LogEventEscalated         LogEvent = "escalated"
)

// A Log is a recording of an Alert event.
type Log struct {
	Timestamp time.Time `json:"timestamp"`
	Event     LogEvent  `json:"event"`
	Message   string    `json:"message"`
}

// Scan handles reading a Role from the DB format
func (r *LogEvent) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*r = LogEvent(t)
	case string:
		*r = LogEvent(t)
	default:
		return fmt.Errorf("could not process unknown type for role %T", t)
	}

	return nil
}
