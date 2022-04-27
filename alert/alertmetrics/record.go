package alertmetrics

import (
	"time"
)

// A Record is a recording of an Alert metric.
type Record struct {
	AlertID   int       
	ServiceID string    
	ClosedAt  time.Time 
	TimeToAck time.Duration
	TimeToClose time.Duration
}
