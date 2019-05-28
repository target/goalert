package lifecycle

// Pausable is able to indicate if a pause operation is on-going.
//
// It is used in cases to initiate a graceful/safe abort of long-running operations
// when IsPausing returns true.
type Pausable interface {
	IsPausing() bool

	// PauseWait will block until a pause operation begins.
	//
	// It should only be used once, it will not block again
	// once resume is called.
	PauseWait() <-chan struct{}
}
