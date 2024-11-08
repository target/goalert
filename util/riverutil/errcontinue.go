package riverutil

import "errors"

// ErrRunAgain is a sentinel error value that can be returned from a function to indicate that the work function should be run again.
//
// This is used when a worker function does partial work, but needs to be run again to complete the work.
var ErrRunAgain = errors.New("run-again")
