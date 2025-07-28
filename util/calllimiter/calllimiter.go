package calllimiter

import (
	"context"
	"sync/atomic"
)

type CallLimiter struct {
	allowed int64
	c       chan struct{}
}

func (c *CallLimiter) Allow() bool {
	if c == nil {
		return true // no limit
	}
	c.c <- struct{}{}
	return atomic.AddInt64(&c.allowed, -1) >= 0
}

func (c *CallLimiter) Release() {
	if c == nil {
		return // no limit
	}
	<-c.c // release the slot
}

type callLimiterContextKey struct{}

func NewCallLimiter(totalLimit, concurrent int) *CallLimiter {
	return &CallLimiter{
		allowed: int64(totalLimit),
		c:       make(chan struct{}, concurrent),
	}
}

func WasLimited(ctx context.Context) bool {
	ql, ok := ctx.Value(callLimiterContextKey{}).(*CallLimiter)
	if !ok {
		return false
	}

	return atomic.LoadInt64(&ql.allowed) < 0
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
