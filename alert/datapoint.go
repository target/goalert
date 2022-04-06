package alert

import (
	"time"
)

// A MetricRecord is a recording of an Alert metric.
type MetricRecord struct {
	AlertID   int       `json:"alert_id"`
	ServiceID string    `json:"service_id"`
	ClosedAt  time.Time `json:"closed_at"`
}
