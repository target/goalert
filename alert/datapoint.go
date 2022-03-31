package alert

import (
	"time"

	"github.com/target/goalert/util/timeutil"
)

// A DataPoint is a recording of an Alert metric.
type DataPoint struct {
	ID          int                  `json:"id"`
	AlertID     int                  `json:"alert_id"`
	ServiceID   string               `json:"service_id"`
	TimeToAck   timeutil.ISODuration `json:"time_to_ack"`
	TimeToClose timeutil.ISODuration `json:"time_to_close"`
	Escalated   bool                 `json:"escalated"`
	Timestamp   time.Time            `json:"timestamp"`
}
