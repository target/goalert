package processinglock

import (
	"github.com/pkg/errors"
)

// Static errors
var (
	// ErrNoLock is returned when a lock can not be acquired due to normal causes.
	ErrNoLock = errors.New("advisory lock already taken or incompatible version")
)
