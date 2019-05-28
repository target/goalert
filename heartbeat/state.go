package heartbeat

import "fmt"

// State represents the health of a heartbeat monitor.
type State string

const (
	// StateInactive means the heartbeat has not yet reported for the first time.
	StateInactive State = "inactive"

	// StateHealthy indicates a heartbeat was received within the past interval.
	StateHealthy State = "healthy"

	// StateUnhealthy indicates a heartbeat has not been received since beyond the interval.
	StateUnhealthy State = "unhealthy"
)

// Scan handles reading State from the DB format
func (r *State) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*r = State(t)
	case string:
		*r = State(t)
	default:
		return fmt.Errorf("could not process unknown type for state %T", t)
	}

	return nil
}
