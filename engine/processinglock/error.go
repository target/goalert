package processinglock

import (
	"github.com/pkg/errors"
)

// Static errors
var (
	// ErrNoLock is returned when a lock can not be acquired due to normal causes.
	ErrNoLock = errors.New("advisory lock already taken or incompatible version")
)

// IsNoLock reports whether err is, or wraps, ErrNoLock.
//
// It is intended for callers (such as River workers) that want to treat lock
// contention as a transient condition, e.g. by snoozing the job, rather than
// as a failure to be logged.
func IsNoLock(err error) bool { return errors.Is(err, ErrNoLock) }
