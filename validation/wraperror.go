package validation

type wrappedError struct {
	err error
}

// WrapError will return a new error that is reported as a ClientError.
func WrapError(err error) error {
	if err == nil {
		return nil
	}
	return &wrappedError{err: err}
}

func (w *wrappedError) ClientError() bool { return true }
func (w *wrappedError) Unwrap() error     { return w.err }
func (w *wrappedError) Error() string     { return w.err.Error() }
