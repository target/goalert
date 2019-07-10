package app

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

var errQueueFull = errors.New("queue full")

// concurrencyLimiter is a locking mechanism that allows setting a limit
// on the number of simultaneous locks for a given arbitrary ID.
type concurrencyLimiter struct {
	maxLock    int
	maxQueue   int
	lockCount  map[string]int
	queueCount map[string]int
	pending    map[string]*pendingList

	mx sync.Mutex
}
type pendingReq struct {
	lockCh     chan bool
	prev, next *pendingReq
}
type pendingList struct{ head, tail *pendingReq }

func newConcurrencyLimiter(maxLock, maxQueue int) *concurrencyLimiter {
	return &concurrencyLimiter{
		maxLock:    maxLock,
		maxQueue:   maxQueue,
		lockCount:  make(map[string]int, 100),
		queueCount: make(map[string]int, 100),
		pending:    make(map[string]*pendingList, 100),
	}
}

// Lock will acquire a lock for the given ID. It may return an err
// if the context expires before the lock is given.
func (l *concurrencyLimiter) Lock(ctx context.Context, id string) (err error) {
	l.mx.Lock()
	c := l.lockCount[id]
	if c < l.maxLock {
		l.lockCount[id] = c + 1
		l.mx.Unlock()
		return nil
	}

	list := l.pending[id]
	if list == nil {
		list = &pendingList{}
		l.pending[id] = list
	}

	if l.queueCount[id] == l.maxQueue {
		l.mx.Unlock()
		return errQueueFull
	}

	// need to queue
	req := &pendingReq{
		lockCh: make(chan bool),
		prev:   list.tail,
	}
	if list.head == nil {
		list.head, list.tail = req, req
	} else {
		list.tail.next = req
		list.tail = req
	}
	l.queueCount[id]++
	l.mx.Unlock()

	cancel := func() {
		close(req.lockCh)
		l.mx.Lock()
		if req.prev != nil {
			req.prev.next = req.next
		}
		if req.next != nil {
			l.queueCount[id]--
			req.next.prev = req.prev
		}
		if list.head == req {
			list.head = req.next
		}
		if list.tail == req {
			list.tail = req.prev
		}
		l.mx.Unlock()
	}
	select {
	case <-ctx.Done():
		cancel()
		return ctx.Err()
	case req.lockCh <- true:
	}

	return nil
}

// Unlock releases a lock for the given ID. It panics
// if there are no remaining locks.
func (l *concurrencyLimiter) Unlock(id string) {
	l.mx.Lock()
	if l.lockCount[id] == 0 {
		l.mx.Unlock()
		panic("not locked")
	}
	list := l.pending[id]
	if list == nil || list.head == nil {
		l.lockCount[id]--
		l.mx.Unlock()
		return
	}

	for list.head != nil {
		head := list.head
		list.head = head.next
		list.head.prev = nil
		head.next = nil
		l.queueCount[id]--
		if <-head.lockCh {
			// keep lock count
			l.mx.Unlock()
			return
		}
	}
	l.lockCount[id]--
	l.mx.Unlock()
}
