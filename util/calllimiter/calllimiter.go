package calllimiter

import (
	"context"
	"sync/atomic"
)

// CallLimiter tracks the number of calls and enforces a maximum limit.
type CallLimiter struct {
	maxCalls int64
	numCalls int64
}

// NumCalls returns the current number of calls made.
func (c *CallLimiter) NumCalls() int {
	if c == nil {
		return 0 // no limit
	}
	return int(atomic.LoadInt64(&c.numCalls))
}

// Allow increments the call count and returns false if the limit is exceeded.
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

// NewCallLimiter creates a new CallLimiter with the specified total limit.
func NewCallLimiter(totalLimit int) *CallLimiter {
	return &CallLimiter{
		maxCalls: int64(totalLimit),
	}
}

// WasLimited checks if the call limit was exceeded and returns the current call count.
func WasLimited(ctx context.Context) (bool, int) {
	ql, ok := ctx.Value(callLimiterContextKey{}).(*CallLimiter)
	if !ok {
		return false, -1
	}
	num := atomic.LoadInt64(&ql.numCalls)
	return num > ql.maxCalls, int(num)
}

// CallLimiterContext creates a new context with a CallLimiter attached.
func CallLimiterContext(ctx context.Context, totalLimit int) context.Context {
	return context.WithValue(ctx, callLimiterContextKey{}, NewCallLimiter(totalLimit))
}

// FromContext retrieves the CallLimiter from the context, or nil if not present.
func FromContext(ctx context.Context) *CallLimiter {
	ql, ok := ctx.Value(callLimiterContextKey{}).(*CallLimiter)
	if !ok {
		return nil
	}
	return ql
}
