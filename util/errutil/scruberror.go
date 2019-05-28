package errutil

import (
	"github.com/pkg/errors"
)

// SafeError is an error string, safe to return to the client.
type SafeError string

// ClientError always returns true.
func (SafeError) ClientError() bool { return true }

func (err SafeError) Error() string { return string(err) }

// ScrubError will replace an err with a generic one if it is not a validation error.
// The boolean value indicates if the error was scrubbed (replaced with a safe one).
func ScrubError(err error) (bool, error) {
	if err == nil {
		return false, nil
	}
	type safe interface {
		ClientError() bool
	}

	if c, ok := errors.Cause(err).(safe); ok && c.ClientError() {
		return false, err
	}

	return true, errors.New("unexpected error")
}
