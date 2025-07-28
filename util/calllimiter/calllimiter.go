package calllimiter

import (
	"context"
	"sync"
)

type CallLimiter struct {
	c *sync.Cond

	maxCalls int
	numCalls int

	maxActive int
	numActive int
}

func (c *CallLimiter) Allow() bool {
	if c == nil {
		return true // no limit
	}
	c.c.L.Lock()
	defer c.c.L.Unlock()
	c.numCalls++
	c.c.Broadcast()
	for {
		if c.numCalls > c.maxCalls {
			return false
		}
		if c.numActive < c.maxActive {
			c.numActive++
			c.c.Broadcast()
			return true
		}
		c.c.Wait()
	}
}

func (c *CallLimiter) Release() {
	if c == nil {
		return // no limit
	}
	c.c.L.Lock()
	defer c.c.L.Unlock()
	c.numActive--
	c.c.Broadcast()
}

type callLimiterContextKey struct{}

func NewCallLimiter(totalLimit, concurrent int) *CallLimiter {
	return &CallLimiter{
		c:         sync.NewCond(&sync.Mutex{}),
		maxCalls:  totalLimit,
		maxActive: concurrent,
	}
}

func WasLimited(ctx context.Context) (bool, int) {
	ql, ok := ctx.Value(callLimiterContextKey{}).(*CallLimiter)
	if !ok {
		return false, -1
	}

	ql.c.L.Lock()
	defer ql.c.L.Unlock()

	return ql.numCalls > ql.maxCalls, ql.numCalls
}

func CallLimiterContext(ctx context.Context, totalLimit, concurrent int) context.Context {
	return context.WithValue(ctx, callLimiterContextKey{}, NewCallLimiter(totalLimit, concurrent))
}

func FromContext(ctx context.Context) *CallLimiter {
	ql, ok := ctx.Value(callLimiterContextKey{}).(*CallLimiter)
	if !ok {
		return nil
	}
	return ql
}
