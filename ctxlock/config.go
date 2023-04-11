package ctxlock

import "time"

// Config is the configuration for an IDLocker.
type Config struct {
	// MaxHeld indicates the maximum number of locks that can be held at once,
	// per ID.
	//
	// If MaxHeld is not set, it defaults to 1.
	MaxHeld int

	// MaxWait indicates the maximum number of pending locks that can be
	// queued for a given ID.
	//
	// If MaxWait is -1, no limit is enforced.
	MaxWait int

	// Timeout indicates the maximum amount of time to wait for a lock to be
	// aquired.
	//
	// If Timeout is 0, no timeout is enforced.
	Timeout time.Duration
}
