package ctxlock

import (
	"errors"
	"sync"
)

var (
	// ErrQueueFull is returned when the queue is full and a lock is requested.
	ErrQueueFull = errors.New("queue full")

	// ErrTimeout is returned when the lock request times out while waiting in
	// the queue.
	ErrTimeout = errors.New("timeout")
)

// IDLocker allows multiple locks to be held at once, but only up to a certain
// number of locks per ID.
//
// If the number of locks for an ID exceeds the maximum, the lock will be
// queued until a lock is released.
//
// It is safe to use IDLocker from multiple goroutines and is used to manage
// concurrency.
type IDLocker[K comparable] struct {
	cfg   Config
	count map[K]int
	queue map[K][]chan struct{}
	mx    sync.Mutex
}

// NewIDLocker creates a new IDLocker with the given config.
//
// An empty config will result in a locker that allows only one lock to be held
// at a time, with no queue (i.e., ErrQueueFull will be returned if a lock is
// requested while another is held).
func NewIDLocker[K comparable](cfg Config) *IDLocker[K] {
	if cfg.MaxHeld == 0 {
		cfg.MaxHeld = 1
	}
	if cfg.MaxHeld < 0 {
		panic("MaxHeld must be >= 0")
	}

	return &IDLocker[K]{
		count: make(map[K]int),
		queue: make(map[K][]chan struct{}),
		cfg:   cfg,
	}
}
