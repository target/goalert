package sendit

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
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

	pending int32
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
	// check done first
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
		_ = ym.Close()
		log.Println("ERROR:", err)
		return
	}

	sess.ym = ym
	close(sess.readyCh)
}

func (sess *session) OpenContext(ctx context.Context) (net.Conn, error) {
	sess.mx.Lock()
	defer sess.mx.Unlock()

	// check done/canceled first
	select {
	case <-sess.doneCh:
		return nil, io.ErrClosedPipe
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

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
		return
	default:
		close(sess.doneCh)
	}

	defer log.Printf("Session ended; %s [%s]", sess.Prefix, sess.ID)

	sess.start.Do(sess.init)
	if sess.ym != nil {
		_ = sess.ym.Close()
	}

	_ = sess.stream.Close()
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

func (sess *session) done() {
	if atomic.AddInt32(&sess.pending, -1) == 0 {
		sess.End()
	}
}

// UseWriter will block until the underlying stream has called Close()
// or the context expires.
func (sess *session) UseWriter(ctx context.Context, w io.Writer) {
	atomic.AddInt32(&sess.pending, 1)
	defer sess.done()

	var rCtx *ioContext
	select {
	case rCtx = <-sess.reader:
	case <-ctx.Done():
		return
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wCtx := &ioContext{
		fn:     w.Write,
		cancel: cancel,
	}

	err := sess.stream.SetPipe(rCtx, wCtx)
	if err != nil {
		log.Println("ERROR: set pipe:", err)
		rCtx.cancel()
	}
	sess.start.Do(sess.init)

	<-ctx.Done()
}

// UseReader will block until the underlying stream has received an io.EOF from the reader
// or the context expires.
func (sess *session) UseReader(ctx context.Context, r io.Reader) {
	atomic.AddInt32(&sess.pending, 1)
	defer sess.done()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ioCtx := &ioContext{
		fn:     r.Read,
		cancel: cancel,
	}
	sess.reader <- ioCtx

	<-ctx.Done()
}
