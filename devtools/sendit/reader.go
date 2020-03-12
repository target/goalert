package sendit

import (
	"context"
	"io"
)

// Reader allows atomically and dynamically chaining multiple io.Reader interfaces.
type Reader struct {
	readerCh chan io.Reader
	eofCh    chan struct{}
	readyCh  chan struct{}
	closeCh  chan struct{}
	err      error
}

// NewReader will create a new Reader.
func NewReader() *Reader {
	return &Reader{
		readerCh: make(chan io.Reader, 1),
		eofCh:    make(chan struct{}, 1),
		readyCh:  make(chan struct{}),
		closeCh:  make(chan struct{}),
	}
}

// Ready returns a channel that will indicate when the reader
// has been activated (first reader set).
//
// It is unaffected by Close()
func (r *Reader) Ready() <-chan struct{} { return r.readyCh }

// Close closes the Reader. Subsequent calls to Read will return `io.EOF`.
//
// It does not call Close on the underlying io.Reader.
func (r *Reader) Close() error {
	select {
	case <-r.closeCh:
		return io.ErrClosedPipe
	default:
	}

	r.err = io.EOF
	close(r.closeCh)

	return nil
}

// CloseWithError closes the Reader. Subsequent calls to Read will return the provided error.
//
// It does not call Close on the underlying io.Reader.
func (r *Reader) CloseWithError(err error) error {
	select {
	case <-r.closeCh:
		return io.ErrClosedPipe
	default:
	}

	r.err = err
	close(r.closeCh)
	return nil
}

// Read will read from the active reader.
func (r *Reader) Read(p []byte) (int, error) {
	var reader io.Reader
	select {
	case <-r.closeCh:
		return 0, r.err
	default:
	}

	select {
	case reader = <-r.readerCh:
	case <-r.closeCh:
		return 0, r.err
	}

	n, err := reader.Read(p)
	if err == nil {
		sr.readerCh <- reader
		return n, err
	}
	if err != io.EOF {
		r.CloseWithError(err)
		return n, err
	}

	sr.eofCh <- struct{}{}

	if n == 0 && len(p) > 0 {
		// nothing read, so try again
		return r.Read(p)
	}

	return n, nil
}

// SetReader will set the active reader to `newR`. If there is no active reader
// then it is set immediately. If a reader is already active, SetReader will
// return after the current reader reaches EOF.
func (r *Reader) SetReader(ctx context.Context, newR io.Reader) error {
	select {
	case <-r.closeCh:
		return io.ErrClosedPipe
	default:
	}

	select {
	case <-r.ready:
	default:
		// not ready, set immediately
		r.readerCh <- newR
		close(r.ready)
		return nil
	}

	select {
	case <-r.closeCh:
		return io.ErrClosedPipe
	case <-ctx.Done():
		return ctx.Err()
	case <-r.eofCh:
		r.readerCh <- newR
		return nil
	}
}
