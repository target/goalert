package permission

import "github.com/pkg/errors"

// Error represents an auth error where the context does not have
// a sufficient role for the operation.
type Error interface {
	error
	Permission() bool // Is the error permission denied?
	Unauthorized() bool
}
type genericError struct {
	unauthorized bool
	reason       string
	stack        errors.StackTrace
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func newGeneric(unauth bool, reason string) genericError {
	return genericError{
		unauthorized: unauth,
		reason:       reason,
		stack:        errors.New("").(stackTracer).StackTrace()[1:],
	}
}

// NewAccessDenied will return a new generic access denied error.
func NewAccessDenied(reason string) error {
	return newGeneric(false, reason)
}

// Unauthorized will return an unauthorized error.
func Unauthorized() error {
	return newGeneric(true, "")
}

func (e genericError) ClientError() bool { return true }

func (e genericError) Permission() bool   { return true }
func (e genericError) Unauthorized() bool { return e.unauthorized }
func (e genericError) Error() string {
	prefix := "access denied"
	if e.unauthorized {
		prefix = "unauthorized"
	}
	if e.reason == "" {
		return prefix
	}

	return prefix + ": " + e.reason
}

// IsPermissionError will determine if the root error cause is a permission error.
func IsPermissionError(err error) bool {
	var e Error
	if errors.As(err, &e) && e.Permission() {
		return true
	}
	return false
}

// IsUnauthorized will determine if the root error cause is an unauthorized permission error.
func IsUnauthorized(err error) bool {
	var e Error
	if errors.As(err, &e) && e.Permission() && e.Unauthorized() {
		return true
	}
	return false
}
