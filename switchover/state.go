package switchover

// State indicates the current state of a node.
type State string

// Possible states
const (
	StateStarting  = State("starting")
	StateReady     = State("ready")
	StateArmed     = State("armed")
	StateArmWait   = State("armed-waiting")
	StatePausing   = State("pausing")
	StatePaused    = State("paused")
	StatePauseWait = State("paused-waiting")
	StateComplete  = State("complete")
	StateAbort     = State("aborted")
)

// IsActive will return true if the state represents
// an on-going change-over event.
func (s State) IsActive() bool {
	switch s {
	case StateArmed, StateArmWait, StatePausing, StatePaused, StatePauseWait:
		return true
	}
	return false
}
