package app

import (
	"context"
	"sync"
)

// concurrencyLimiter is a locking mechanism that allows setting a limit
// on the number of simultaneous locks for a given arbitrary ID.
type concurrencyLimiter struct {
	max     int
	count   map[string]int
	pending map[string]chan struct{}

	mx sync.Mutex
}

func newConcurrencyLimiter(max int) *concurrencyLimiter {
	return &concurrencyLimiter{
		max:     max,
		count:   make(map[string]int, 100),
		pending: make(map[string]chan struct{}),
	}
}

// Lock will acquire a lock for the given ID. It may return an err
// if the context expires before the lock is given.
func (l *concurrencyLimiter) Lock(ctx context.Context, id string) error {
	for {
		l.mx.Lock()
		n := l.count[id]
		if n < l.max {
			l.count[id] = n + 1
			l.mx.Unlock()
			return nil
		}

		ch := l.pending[id]
		if ch == nil {
			ch = make(chan struct{})
			l.pending[id] = ch
		}
		l.mx.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
		}
	}
}

// Unlock releases a lock for the given ID. It panics
// if there are no remaining locks.
func (l *concurrencyLimiter) Unlock(id string) {
	l.mx.Lock()
	defer l.mx.Unlock()
	n := l.count[id]
	n--
	if n < 0 {
		panic("not locked: " + id)
	}
	if n == 0 {
		delete(l.count, id)
	} else {
		l.count[id] = n
	}
	ch := l.pending[id]
	if ch != nil {
		delete(l.pending, id)
		close(ch)
	}
}
