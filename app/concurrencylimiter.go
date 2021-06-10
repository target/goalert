package app

import (
	"container/list"
	"context"
	"sync"

	"github.com/pkg/errors"
)

var errQueueFull = errors.New("queue full")

type multiLock struct {
	queue *list.List

	lockCount int
	maxLock   int
	maxWait   int

	lock     chan struct{}
	lockFull chan struct{}
}

func newMultiLock(maxLock, maxWait int) *multiLock {
	m := &multiLock{
		maxLock:  maxLock,
		maxWait:  maxWait,
		queue:    list.New(),
		lock:     make(chan struct{}, 1),
		lockFull: make(chan struct{}, 1),
	}
	m.lock <- struct{}{}
	return m
}

func (m *multiLock) Lock(ctx context.Context) (err error) {
	select {
	case <-m.lock:
		// fast path if not yet at the limit
		m.lockCount++
		if m.lockCount == m.maxLock {
			m.lockFull <- struct{}{}
		} else {
			m.lock <- struct{}{}
		}
		return nil
	case <-m.lockFull:
	}

	if m.queue.Len() == m.maxWait {
		m.lockFull <- struct{}{}
		return errQueueFull
	}

	lockCh := make(chan struct{})
	e := m.queue.PushBack(lockCh)
	m.lockFull <- struct{}{}

	select {
	case <-lockCh:
		return nil
	case <-ctx.Done():
	}

	// canceled; attempt to remove or accept lock
	select {
	case <-lockCh:
		// got lock, return nil
		return nil
	case <-m.lockFull:
	}

	m.queue.Remove(e)
	m.lockFull <- struct{}{}

	return ctx.Err()
}
func (m *multiLock) Unlock() (last bool) {
	select {
	case <-m.lock:
		// no queue, just decrement
		if m.lockCount == 0 {
			m.lock <- struct{}{}
			panic("not locked")
		}
		m.lockCount--
		last = m.lockCount == 0
		m.lock <- struct{}{}
		return last
	case <-m.lockFull:
	}

	if m.queue.Len() == 0 {
		// no queue, no longer full
		m.lockCount--
		last = m.lockCount == 0
		m.lock <- struct{}{}
		return last
	}

	m.queue.Remove(m.queue.Front()).(chan struct{}) <- struct{}{}
	m.lockFull <- struct{}{}

	return false
}

// concurrencyLimiter is a locking mechanism that allows setting a limit
// on the number of simultaneous locks for a given arbitrary ID.
type concurrencyLimiter struct {
	maxLock int
	maxWait int

	locks map[string]*multiLock
	lruEl map[string]*list.Element

	lruEmpty *list.List

	mx sync.Mutex
}

func newConcurrencyLimiter(maxLock, maxWait int) *concurrencyLimiter {
	return &concurrencyLimiter{
		maxLock:  maxLock,
		maxWait:  maxWait,
		lruEmpty: list.New(),
		locks:    make(map[string]*multiLock, 100),
		lruEl:    make(map[string]*list.Element, 100),
	}
}

// Lock will acquire a lock for the given ID. It may return an err
// if the context expires before the lock is given.
func (l *concurrencyLimiter) Lock(ctx context.Context, id string) (err error) {
	l.mx.Lock()
	m := l.locks[id]
	if m == nil {
		m = newMultiLock(l.maxLock, l.maxWait)
		l.locks[id] = m
	}

	// Remove an empty entry if it exists.
	// Lock is guaranteed to succeed if empty (making it non-empty).
	el := l.lruEl[id]
	if el != nil {
		delete(l.lruEl, id)
		l.lruEmpty.Remove(el)
	}
	l.mx.Unlock()

	return m.Lock(ctx)
}

// Unlock releases a lock for the given ID.
func (l *concurrencyLimiter) Unlock(id string) {
	l.mx.Lock()
	defer l.mx.Unlock()

	empty := l.locks[id].Unlock()
	if !empty {
		// nothing to cleanup, not empty
		return
	}

	// push the ID to the front of the LRU list.
	l.lruEl[id] = l.lruEmpty.PushFront(id)

	if l.lruEmpty.Len() < 100 {
		// do nothing if we have less than 100 in memory
		return
	}

	// remove oldest ID from memory
	purgeID := l.lruEmpty.Remove(l.lruEmpty.Back()).(string)
	delete(l.locks, purgeID)
	delete(l.lruEl, purgeID)
}
