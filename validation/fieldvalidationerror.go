package validation

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// A FieldError represents an invalid field during validation.
type FieldError interface {
	error
	Validation() bool
	Field() string
	Reason() string
}
type MultiFieldError interface {
	error
	Validation() bool
	FieldErrors() []FieldError
}

type validation interface {
	Validation() bool
}
type client interface {
	ClientError() bool
}

type fieldError struct {
	stack     errors.StackTrace
	reason    string
	fieldName string
}
type fieldErrors struct {
	stack  errors.StackTrace
	errors []FieldError
}

// AddPrefix will prepend a prefix to all field names within the given error.
func AddPrefix(fieldPrefix string, err error) error {
	switch e := errors.Cause(err).(type) {
	case *fieldError:
		e.fieldName = fieldPrefix + e.fieldName
	case *fieldErrors:
		for _, err := range e.errors {
			if e, ok := err.(*fieldError); ok {
				e.fieldName = fieldPrefix + e.fieldName
			}
		}
	}
	return err
}

func (f fieldError) ClientError() bool  { return true }
func (f fieldErrors) ClientError() bool { return true }

func (f fieldError) StackTrace() errors.StackTrace { return f.stack }
func (f fieldError) Reason() string                { return f.reason }
func (f fieldError) Error() string {
	return fmt.Sprintf("invalid value for '%s': %s", f.fieldName, f.reason)
}
func (f fieldError) Validation() bool { return true }
func (f fieldError) Field() string    { return f.fieldName }

func (f fieldErrors) Validation() bool { return true }
func (f fieldErrors) Field() string {
	names := make([]string, len(f.errors))
	for i, e := range f.errors {
		names[i] = e.Field()
	}
	return strings.Join(names, ",")
}
func (f fieldErrors) StackTrace() errors.StackTrace { return f.stack }
func (f fieldErrors) Error() string {
	strs := make([]string, len(f.errors))
	for i, e := range f.errors {
		strs[i] = e.Error()
	}
	return strings.Join(strs, "\n")
}
func (f fieldErrors) FieldErrors() []FieldError {
	errs := make([]FieldError, len(f.errors))
	copy(errs, f.errors)
	return errs
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// NewMultiFieldError will combine multiple FieldErrors into a MultiFieldError.
func NewMultiFieldError(errs []FieldError) MultiFieldError {
	return &fieldErrors{errors: errs, stack: errors.New("").(stackTracer).StackTrace()}
}

// NewFieldError will create a new FieldError for the given field and reason
func NewFieldError(fieldName string, reason string) FieldError {
	return &fieldError{reason: reason, fieldName: fieldName, stack: errors.New("").(stackTracer).StackTrace()}
}

// IsValidationError will determine if an error's cause is a field validation error.
func IsValidationError(err error) bool {
	if e, ok := errors.Cause(err).(validation); ok && e.Validation() {
		return true
	}
	return false
}

// IsClientError will determine if an error's cause is due to request/client error.
func IsClientError(err error) bool {
	if e, ok := errors.Cause(err).(client); ok && e.ClientError() {
		return true
	}
	return false
}
