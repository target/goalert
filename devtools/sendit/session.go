package sendit

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

type session struct {
	ID     string
	Prefix string

	stream *Stream
	ym     *yamux.Session
	s      *Server

	mx sync.Mutex

	start sync.Once

	reader chan *ioContext

	doneCh  chan struct{}
	readyCh chan struct{}

	wg sync.WaitGroup
}

func (s *Server) newSession(prefix string) (*session, error) {
	if !ValidPath(prefix) {
		return nil, errors.New("invalid prefix value")
	}
	id, err := genID()
	if err != nil {
		panic(err)
	}

	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.sessionsByPrefix[prefix]; ok {
		return nil, errors.New("prefix already in use")
	}

	sess := &session{
		ID:     id,
		Prefix: prefix,
		s:      s,

		stream:  NewStream(),
		reader:  make(chan *ioContext, 1),
		doneCh:  make(chan struct{}),
		readyCh: make(chan struct{}),
	}

	s.sessionsByPrefix[prefix] = sess
	s.sessionsByID[id] = sess

	return sess, nil
}

func (sess *session) init() {
	select {
	case <-sess.doneCh:
		return
	default:
	}

	select {
	case <-sess.stream.Ready():
	case <-sess.doneCh:
		return
	}

	cfg := yamux.DefaultConfig()
	cfg.KeepAliveInterval = 3 * time.Second
	ym, err := yamux.Server(sess.stream, cfg)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}

	_, err = ym.Ping()
	if err != nil {
		ym.Close()
		log.Println("ERROR:", err)
		return
	}

	sess.ym = ym
	close(sess.readyCh)
	go func() {
		sess.wg.Wait()
		sess.End()
	}()
}

func (sess *session) OpenContext(ctx context.Context) (net.Conn, error) {
	select {
	case <-sess.doneCh:
		return nil, io.ErrClosedPipe
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-sess.readyCh:
	}

	return sess.ym.Open()
}

func (sess *session) End() {
	sess.mx.Lock()
	defer sess.mx.Unlock()

	select {
	case <-sess.doneCh:
	default:
		close(sess.doneCh)
	}

	sess.start.Do(sess.init)
	if sess.ym != nil {
		sess.ym.Close()
	}

	sess.stream.Close()
	sess.s.mx.Lock()
	defer sess.s.mx.Unlock()

	delete(sess.s.sessionsByID, sess.ID)
	delete(sess.s.sessionsByPrefix, sess.Prefix)
}

type ioContext struct {
	fn     func([]byte) (int, error)
	cancel func()
}

func (ctx *ioContext) Read(p []byte) (int, error)  { return ctx.fn(p) }
func (ctx *ioContext) Write(p []byte) (int, error) { return ctx.fn(p) }
func (ctx *ioContext) Close() error                { ctx.cancel(); return nil }

// UseWriter will block until the underlying stream has called Close()
// or the context expires.
func (sess *session) UseWriter(ctx context.Context, w io.Writer) {
	sess.wg.Add(1)
	defer sess.wg.Done()

	var rCtx *ioContext
	select {
	case rCtx = <-sess.reader:
	case <-ctx.Done():
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wCtx := &ioContext{
		fn: w.Write,
		cancel: func() {
			w.(http.Flusher).Flush() // ensure we explicitly flush before returning
			cancel()
		},
	}

	err := sess.stream.SetPipe(rCtx, wCtx)
	if err != nil {
		rCtx.cancel()
	}
	sess.start.Do(sess.init)

	<-ctx.Done()
}

// UseReader will block until the underlying stream has received an io.EOF from the reader
// or the context expires.
func (sess *session) UseReader(ctx context.Context, r io.Reader) {
	sess.wg.Add(1)
	defer sess.wg.Done()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ioCtx := &ioContext{
		fn: func(p []byte) (n int, err error) {
			n, err = r.Read(p)
			if err == io.EOF {
				cancel()
			}
			return n, err
		},
		cancel: cancel,
	}
	sess.reader <- ioCtx

	<-ctx.Done()
}
