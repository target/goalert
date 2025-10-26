package sendit

import (
	"bufio"
	"errors"
	"io"
	"sync"
)

// Stream is a ReadWriteCloser that can have it's read/write pipe replaced safely while actively sending data.
type Stream struct {
	readCh    chan io.ReadCloser
	readEOFCh chan io.ReadCloser
	writeCh   chan io.WriteCloser
	closeCh   chan struct{}
	mx        sync.Mutex

	readyCh chan struct{}

	isClosed bool
	isReady  bool
}

// NewStream initializes a new stream. SetPipe must be called before data can be transferred.
func NewStream() *Stream {
	s := &Stream{
		readCh:    make(chan io.ReadCloser, 1),
		readEOFCh: make(chan io.ReadCloser, 1),
		writeCh:   make(chan io.WriteCloser, 1),
		closeCh:   make(chan struct{}),
		readyCh:   make(chan struct{}),
	}
	return s
}

// Ready will return a channel indicating when the stream is ready for reading and writing.
func (s *Stream) Ready() <-chan struct{} { return s.readyCh }

func (s *Stream) Write(p []byte) (int, error) {
	var w io.WriteCloser
	select {
	case w = <-s.writeCh:
	case <-s.closeCh:
		return 0, io.ErrClosedPipe
	}

	n, err := w.Write(p)
	s.writeCh <- w
	return n, err
}

func (s *Stream) Read(p []byte) (int, error) {
	var r io.ReadCloser
	select {
	case r = <-s.readCh:
	case <-s.closeCh:
		return 0, io.EOF
	}

	n, err := r.Read(p)
	if err == io.EOF {
		s.readEOFCh <- r
		if n == 0 {
			return s.Read(p)
		}
	} else {
		s.readCh <- r
	}
	return n, err
}

// Close will shutdown the stream, and close the underlying Writer.
func (s *Stream) Close() (err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if s.isClosed {
		return nil
	}

	s.isClosed = true
	if !s.isReady {
		close(s.closeCh)
		return nil
	}

	select {
	case wc := <-s.writeCh:
		err = wc.Close()
	case <-s.closeCh:
		return io.ErrClosedPipe
	}
	select {
	case r := <-s.readCh:
		r.Close()
	case r := <-s.readEOFCh:
		r.Close()
	}
	close(s.closeCh)

	return err
}

// SetPipe will swap the io.ReadCloser and WriteCloser with the underlying Stream safely.
func (s *Stream) SetPipe(r io.ReadCloser, wc io.WriteCloser) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.isClosed {
		return io.ErrClosedPipe
	}

	_, err := io.WriteString(wc, "SYN\n")
	if err != nil {
		return err
	}
	var bufClose struct {
		*bufio.Reader
		io.Closer
	}
	bufClose.Reader = bufio.NewReader(r)
	bufClose.Closer = r
	str, err := bufClose.ReadString('\n')
	if err != nil {
		return err
	}
	if str != "SYN\n" {
		return errors.New("expected SYN")
	}

	// Now that Read calls will safely transition to the next pipe, we can give the other end the OK
	// to terminate the first.
	_, err = io.WriteString(wc, "SYNACK\n")
	if err != nil {
		return err
	}
	str, err = bufClose.ReadString('\n')
	if err != nil {
		return err
	}
	if str != "SYNACK\n" {
		return errors.New("expected SYNACK")
	}

	if !s.isReady {
		s.writeCh <- wc
		s.readCh <- bufClose
		s.isReady = true
		close(s.readyCh)
		return nil
	}

	closeOld := func(c io.Closer) bool {
		err = c.Close()
		if err != nil {
			wc.Close()
			r.Close()
			s.isClosed = true
			close(s.closeCh)
			return false
		}
		return true
	}

	// Now we've established the new pipe, we just need to swap and close the read/writers.
	select {
	case oldWriter := <-s.writeCh:
		s.writeCh <- wc
		if !closeOld(oldWriter) {
			return err
		}
		oldReader := <-s.readEOFCh
		s.readCh <- bufClose
		if !closeOld(oldReader) {
			return err
		}
	case oldReader := <-s.readEOFCh:
		s.readCh <- bufClose
		if !closeOld(oldReader) {
			return err
		}
		oldWriter := <-s.writeCh
		s.writeCh <- wc
		if !closeOld(oldWriter) {
			return err
		}
	}

	return nil
}
