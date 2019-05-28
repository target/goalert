package heartbeat

import (
	"database/sql"
	"github.com/target/goalert/validation/validate"
)

// A Monitor will generate an alert if it does not receive a heartbeat within the configured IntervalMinutes.
type Monitor struct {
	ID              string `json:"id,omitempty"`
	Name            string `json:"name,omitempty"`
	ServiceID       string `json:"service_id,omitempty"`
	IntervalMinutes int    `json:"interval_minutes,omitempty"`

	lastState            State
	lastHeartbeatMinutes sql.NullInt64
}

// LastState returns the last known state.
func (m Monitor) LastState() State { return m.lastState }

// LastHeartbeatMinutes returns the minutes since the heartbeat last reported.
// The interval is truncated, so a value of 0 means "less than 1 minute".
func (m Monitor) LastHeartbeatMinutes() (int, bool) {
	return int(m.lastHeartbeatMinutes.Int64), m.lastHeartbeatMinutes.Valid
}

// Normalize performs validation and returns a new copy.
func (m Monitor) Normalize() (*Monitor, error) {
	err := validate.Many(
		validate.UUID("ServiceID", m.ServiceID),
		validate.IDName("Name", m.Name),
		validate.Range("IntervalMinutes", m.IntervalMinutes, 1, 9000),
	)
	if err != nil {
		return nil, err
	}

	return &m, nil
}
