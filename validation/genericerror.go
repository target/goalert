package validation

import "github.com/pkg/errors"

type genericError struct {
	message string
	stack   errors.StackTrace
}

// NewGenericError will create a new generic validation error.
func NewGenericError(message string) error {
	return &genericError{message: message, stack: errors.New("").(stackTracer).StackTrace()}
}

func (g genericError) Error() string                 { return g.message }
func (g genericError) StackTrace() errors.StackTrace { return g.stack }
func (g genericError) ClientError() bool             { return true }
