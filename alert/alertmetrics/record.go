package alertmetrics

import (
	"time"
)

// A Record is a recording of an Alert metric.
type Record struct {
	ServiceID   string
	AlertCount  int
	ClosedAt    time.Time
	TimeToAck   time.Duration
	TimeToClose time.Duration
}
