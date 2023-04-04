package app

import (
	"container/list"
	"context"
	"sync"

	"github.com/pkg/errors"
)

var errQueueFull = errors.New("queue full")

// multiLock allows a fixed number of Lock calls to succeed
// with a maximum ordered-queue.
type multiLock struct {
	queue *list.List

	lockCount int
	maxLock   int
	maxWait   int

	lock     chan struct{}
	lockFull chan struct{}
}

// newMultiLock creates a new multiLock with the provided max lock and wait counts.
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

// TryLock is guaranteed to return immediately. If the request has been put in the queue,
// a waitFn is returned. If err != nil, the lock failed. If both are nil, the lock was
// acquired.
//
// waitFn will return nil if the lock was acquired.
func (m *multiLock) TryLock() (waitFn func(context.Context) error, err error) {
	select {
	case <-m.lock:
		// fast path if not yet at the limit
		m.lockCount++
		if m.lockCount == m.maxLock {
			m.lockFull <- struct{}{}
		} else {
			m.lock <- struct{}{}
		}
		return nil, nil
	case <-m.lockFull:
	}

	if m.queue.Len() == m.maxWait {
		m.lockFull <- struct{}{}
		return nil, errQueueFull
	}

	lockCh := make(chan struct{})
	e := m.queue.PushBack(lockCh)
	m.lockFull <- struct{}{}

	return func(ctx context.Context) error {
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
	}, nil
}

// Lock will return nil if a lock was acquired or errQueueFull if there is no more room
// in the queue. If the context is canceled while waiting in the queue ctx.Err() may be
// returned.
//
// The lock is valid beyond the lifecyle of the provided context, if nil is returned
// Unlock must be called to release it.
//
// The number of locks held is guaranteed to be >0 after calling Lock.
func (m *multiLock) Lock(ctx context.Context) error {
	waitFn, err := m.TryLock()
	if err != nil {
		return err
	}
	if waitFn == nil {
		return nil
	}

	return waitFn(ctx)
}

// Unlock will release a held lock. It returns true
// if the new lock count is zero.
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

	// Use two-stage locking so we can "reserve" our lock before releasing the mutex.
	// This prevents a possible race condition with cleanup thinking the lock is empty
	// before Lock is actually called here.
	//
	// Two stage ensures we either get the lock or are added to the queue (marking it as non-empty)
	// before allowing Unlock to do it's cleanup logic.
	wait, err := m.TryLock()
	l.mx.Unlock()
	if err != nil {
		return err
	}
	if wait == nil {
		return nil
	}

	return wait(ctx)
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
