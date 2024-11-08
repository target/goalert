package cleanupmanager

import (
	"context"
	"time"
)

// runWhileWork will run the provided function until it returns 0 rows or the context is canceled.
func runWhileWork(ctx context.Context, runFn func() (int64, error)) error {
	for {

		rows, err := runFn()
		if err != nil {
			return err
		}
		if rows == 0 {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}
