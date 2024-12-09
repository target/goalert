package util

import (
	"context"
	"time"
)

// ContextSleep will sleep for the specified duration or until the context is canceled.
func ContextSleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
