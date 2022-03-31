package alert

import (
	"time"
)

// A DataPoint is a recording of an Alert metric.
type DataPoint struct {
	AlertID     int                  `json:"alert_id"`
	ServiceID   string               `json:"service_id"`
	Timestamp   time.Time            `json:"timestamp"`
}
