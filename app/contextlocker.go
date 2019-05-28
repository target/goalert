package app

import (
	"context"
	"sync/atomic"
)

type contextLocker struct {
	readCount int64

	lock   chan lockReq
	unlock chan chan struct{}

	rLock      chan struct{}
	rUnlock    chan struct{}
	rNotLocked chan struct{}
}
type lockReq struct {
	cancel <-chan struct{}
	ch     chan bool
}

func newContextLocker() *contextLocker {
	c := &contextLocker{
		lock:       make(chan lockReq),
		unlock:     make(chan chan struct{}, 1),
		rLock:      make(chan struct{}),
		rUnlock:    make(chan struct{}),
		rNotLocked: make(chan struct{}),
	}
	go c.loop()
	return c
}
func (c *contextLocker) writeLock(req lockReq) {
	for atomic.LoadInt64(&c.readCount) > 0 {
		select {
		case <-c.rUnlock:
			atomic.AddInt64(&c.readCount, -1)
		case <-req.cancel:
			req.ch <- false
			return
		}
	}

	ch := make(chan struct{})
	c.unlock <- ch
	req.ch <- true
	for {
		select {
		case <-ch:
			return
		case <-c.rNotLocked:
		}
	}
}

func (c *contextLocker) loop() {
	for {
		select {
		// request for write lock always takes precedence
		case req := <-c.lock:
			c.writeLock(req)
			continue
		default:
		}

		if atomic.LoadInt64(&c.readCount) == 0 {
			select {
			case req := <-c.lock:
				c.writeLock(req)
			case <-c.rLock:
				atomic.AddInt64(&c.readCount, 1)
			case <-c.rNotLocked:
			}
			continue
		}

		select {
		case req := <-c.lock:
			c.writeLock(req)
		case <-c.rLock:
			atomic.AddInt64(&c.readCount, 1)
		case <-c.rUnlock:
			atomic.AddInt64(&c.readCount, -1)
		}
	}
}
func (c *contextLocker) RLockCount() int {
	return int(atomic.LoadInt64(&c.readCount))
}
func (c *contextLocker) Lock(ctx context.Context) error {
	ch := make(chan bool)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.lock <- lockReq{cancel: ctx.Done(), ch: ch}:
	}

	if <-ch {
		return nil
	}

	return ctx.Err()
}
func (c *contextLocker) Unlock() {
	select {
	case ch := <-c.unlock:
		ch <- struct{}{}
	default:
		// safe to call, even if not write-locked (unlike RUnlock)
	}
}
func (c *contextLocker) RLock(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.rLock <- struct{}{}:
	}

	return nil
}
func (c *contextLocker) RUnlock() {
	select {
	case c.rUnlock <- struct{}{}:
	case c.rNotLocked <- struct{}{}:
		panic("not locked")
	}
}
