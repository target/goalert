package calllimiter

import (
	"context"
	"sync/atomic"
)

type CallLimiter struct {
	maxCalls int64
	numCalls int64
}

func (c *CallLimiter) Allow() bool {
	if c == nil {
		return true // no limit
	}
	if atomic.AddInt64(&c.numCalls, 1) > c.maxCalls {
		return false
	}

	return true
}

type callLimiterContextKey struct{}

func NewCallLimiter(totalLimit, concurrent int) *CallLimiter {
	return &CallLimiter{
		maxCalls: int64(totalLimit),
	}
}

func WasLimited(ctx context.Context) (bool, int) {
	ql, ok := ctx.Value(callLimiterContextKey{}).(*CallLimiter)
	if !ok {
		return false, -1
	}
	num := atomic.LoadInt64(&ql.numCalls)
	return num > ql.maxCalls, int(num)
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
