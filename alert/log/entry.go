package alertlog

import (
	"time"
)

// An Entry is an alert log entry.
type Entry interface {
	// AlertID returns the ID of the alert the Entry belongs to.
	AlertID() int

	// ID returns the ID of the log entry.
	ID() int
	// Timestamp returns the time the Entry was created.
	Timestamp() time.Time

	// Type returns type type of log entry.
	Type() Type

	// Subject will return the subject, if available of the Entry.
	Subject() *Subject

	// String returns the string representation of a log Event.
	String() string

	Meta() interface{}
}
