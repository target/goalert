package retry

import (
	"context"
	"time"
)

// Wrap will wrap a single-argument function (excl. context) with retry logic.
func Wrap[T, A any](fn func(context.Context, A) (T, error), attempts int, backoff time.Duration) func(context.Context, A) (T, error) {
	return func(ctx context.Context, a A) (T, error) {
		var res T
		var err error
		err = DoTemporaryError(func(int) error {
			res, err = fn(ctx, a)
			return err
		},
			Log(ctx),
			Limit(attempts),
			FibBackoff(backoff),
		)
		return res, err
	}
}

// Wrap2 will wrap a two-argument function (excl. context) with retry logic.
func Wrap2[T, A, B any](fn func(context.Context, A, B) (T, error), attempts int, backoff time.Duration) func(context.Context, A, B) (T, error) {
	return func(ctx context.Context, a A, b B) (T, error) {
		var res T
		var err error
		err = DoTemporaryError(func(int) error {
			res, err = fn(ctx, a, b)
			return err
		},
			Log(ctx),
			Limit(attempts),
			FibBackoff(backoff),
		)
		return res, err
	}
}
