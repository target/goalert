package sendit

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"io"
	"sync"
)

var key [32]byte

func FlipReader(r io.Reader) io.Reader {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	return &cipher.StreamReader{R: r, S: stream}
}

func FlipWriter(w io.Writer) io.Writer {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		panic(err)
	}
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	return &cipher.StreamWriter{W: w, S: stream}
}

type Stream struct {
	readCh  chan []io.ReadCloser
	writeCh chan io.WriteCloser
	closeCh chan struct{}
	mx      sync.Mutex

	readyCh chan struct{}

	isClosed bool
	isReady  bool
}

func NewStream() *Stream {
	s := &Stream{
		readCh:  make(chan []io.ReadCloser, 1),
		writeCh: make(chan io.WriteCloser, 1),
		closeCh: make(chan struct{}),
		readyCh: make(chan struct{}),
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
	var r []io.ReadCloser
	select {
	case r = <-s.readCh:
	case <-s.closeCh:
		return 0, io.EOF
	}
	if len(r) == 0 {
		s.readCh <- r
		return 0, io.EOF
	}

	n, err := r[0].Read(p)
	if err == io.EOF {
		r[0].Close()
		r = r[1:]
		if n == 0 {
			s.readCh <- r
			return s.Read(p)
		}
	}
	s.readCh <- r
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
	for _, r := range <-s.readCh {
		r.Close()
	}
	close(s.closeCh)

	return err
}

// SetPipe will swap the io.Reader and WriteCloser with the underlying Stream safely.
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
	br := bufio.NewReader(r)
	str, err := br.ReadString('\n')
	if err != nil {
		return err
	}
	if str != "SYN\n" {
		return errors.New("expected SYN")
	}

	var bufCloser struct {
		*bufio.Reader
		io.Closer
	}
	bufCloser.Reader = br
	bufCloser.Closer = r
	// We have SYN so we know the read half is ready, so swap that first so we don't get an unexpected EOF.
	//
	// The reader has to be ready before we send our SYNACK message, which gives the other end the OK
	// to flush and close their current writer (our previous reader).
	if s.isReady {
		s.readCh <- append(<-s.readCh, bufCloser)
	} else {
		s.readCh <- []io.ReadCloser{bufCloser}
	}

	// Now that Read calls will safely transition to the next pipe, we can give the other end the OK
	// to terminate the first.
	_, err = io.WriteString(wc, "SYNACK\n")
	if err != nil {
		return err
	}
	str, err = br.ReadString('\n')
	if err != nil {
		return err
	}
	if str != "SYNACK\n" {
		return errors.New("expected SYNACK")
	}

	// Now both sides know it's OK to flush and terminate the old write connection.
	if s.isReady {
		oldWriter := <-s.writeCh
		err := oldWriter.Close()
		if err != nil {
			// Stream is in an unknown state, we failed to close/flush our end properly, terminate everything
			wc.Close()
			s.isClosed = true
			close(s.closeCh)
			return err
		}
	} else {
		close(s.readyCh)
		s.isReady = true
	}

	s.writeCh <- wc

	return nil
}
