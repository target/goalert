package retry

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
)

var _fib = []int{0, 1}

func fib(n int) int {
	for i := len(_fib) - 1; i < n; i++ {
		_fib = append(_fib, _fib[i-1]+_fib[i])
	}
	return _fib[n]
}

func init() {
	fib(30)
}

// An Option takes the attempt number and the last error value (can be nil) and should indicate
// if a retry should be made.
type Option func(int, error) bool

// Do will retry the given DoFunc until it or an option returns false. The last returned
// error value (can be nil) of fn will be returned.
//
// fn will be passed the current attempt number (starting with 0).
func Do(fn func(attempt int) (shouldRetry bool, err error), opts ...Option) error {
	var n int
	var err error
	var retry bool
	var opt Option
	for {
		for _, opt = range opts {
			if !opt(n, err) {
				return err
			}
		}
		retry, err = fn(n)
		if !retry {
			return err
		}
		n++
	}
}

// Log will log all errors between retries returned from the DoFunc at Debug level. The final error, if any, is not logged.
func Log(ctx context.Context) Option {
	return func(a int, err error) bool {
		if a == 0 || err == nil {
			return true
		}
		log.Debug(log.WithField(ctx, "RetryAttempt", a-1), errors.Wrap(err, "will retry"))
		return true
	}
}

// Limit will set the max number of retry attempts (including the initial attempt).
func Limit(n int) Option {
	return func(a int, _ error) bool {
		return a < n
	}
}

// FibBackoff will Sleep for f(n) * Duration (+/- 50ms) before each attempt, where f(n) is the value from the Fibonacci sequence for
// the nth attempt. There is no delay for the first attempt (n=0).
func FibBackoff(d time.Duration) Option {
	return func(a int, _ error) bool {
		if a == 0 {
			return true
		}
		time.Sleep(time.Duration(fib(a))*d + time.Duration(rand.Intn(100)-50)*time.Millisecond)
		return true
	}
}

// Context will allow retry to continue until the context is canceled.
func Context(ctx context.Context) Option {
	return func(a int, _ error) bool {
		return ctx.Err() == nil
	}
}
