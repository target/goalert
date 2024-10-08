package heartbeat

import (
	"time"

	"github.com/jackc/pgtype"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// A Monitor will generate an alert if it does not receive a heartbeat within the configured TimeoutMinutes.
type Monitor struct {
	ID        string        `json:"id,omitempty"`
	Name      string        `json:"name,omitempty"`
	ServiceID string        `json:"service_id,omitempty"`
	Timeout   time.Duration `json:"timeout,omitempty"`

	AdditionalDetails string

	// Muted indicates the reason the monitor is muted.
	//
	// If non-empty, the monitor will not generate alerts.
	Muted string

	lastState     State
	lastHeartbeat time.Time
}

// LastState returns the last known state.
func (m Monitor) LastState() State { return m.lastState }

// LastHeartbeat returns the timestamp of the last successful heartbeat.
func (m Monitor) LastHeartbeat() time.Time { return m.lastHeartbeat }

// Normalize performs validation and returns a new copy.
func (m Monitor) Normalize() (*Monitor, error) {
	err := validate.Many(
		validate.UUID("ServiceID", m.ServiceID),
		validate.IDName("Name", m.Name),
		validate.Duration("Timeout", m.Timeout, 5*time.Minute, 9000*time.Hour),
		validate.Text("AdditionalDetails", m.AdditionalDetails, 0, alert.MaxDetailsLength),
		validate.Text("Muted", m.Muted, 0, 255),
	)
	if err != nil {
		return nil, err
	}

	m.Timeout = m.Timeout.Truncate(time.Minute)

	return &m, nil
}

func (m *Monitor) scanFrom(scanFn func(...interface{}) error) error {
	var (
		t       sqlutil.NullTime
		timeout pgtype.Interval
	)

	err := scanFn(&m.ID, &m.Name, &m.ServiceID, &timeout, &m.lastState, &t, &m.AdditionalDetails)
	if err != nil {
		return err
	}

	err = timeout.AssignTo(&m.Timeout)
	if err != nil {
		return err
	}

	m.lastHeartbeat = t.Time

	return nil
}
