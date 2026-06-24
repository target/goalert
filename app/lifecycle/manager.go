package lifecycle

import (
	"context"

	"github.com/pkg/errors"
)

// Status represents lifecycle state.
type Status int

// Possible states.
const (
	StatusUnknown Status = iota
	StatusStarting
	StatusReady
	StatusShutdown
	StatusPausing
	StatusPaused
)

// Static errors
var (
	ErrAlreadyStarted   = errors.New("already started")
	ErrShutdown         = errors.New("shutting down")
	ErrNotStarted       = errors.New("not started")
	ErrPauseUnsupported = errors.New("pause not supported or unset")
)

// Manager is used to wrap lifecycle methods with strong guarantees.
type Manager struct {
	startupFunc  func(context.Context) error
	runFunc      func(context.Context) error
	shutdownFunc func(context.Context) error
	pauseResume  PauseResumer

	status chan Status

	startupCancel func()
	startupDone   chan struct{}
	startupErr    error

	runCancel func()
	runDone   chan struct{}

	shutdownCancel func()
	shutdownDone   chan struct{}
	shutdownErr    error

	pauseCancel func()
	pauseDone   chan struct{}
	pauseStart  chan struct{}
	pauseErr    error
	isPausing   bool
}

var (
	_ Pausable     = &Manager{}
	_ PauseResumer = &Manager{}
)

// NewManager will construct a new manager wrapping the provided
// run and shutdown funcs.
func NewManager(run, shutdown func(context.Context) error) *Manager {
	mgr := &Manager{
		runFunc:      run,
		shutdownFunc: shutdown,

		runDone:      make(chan struct{}),
		startupDone:  make(chan struct{}),
		shutdownDone: make(chan struct{}),
		pauseStart:   make(chan struct{}),
		status:       make(chan Status, 1),
	}
	mgr.status <- StatusUnknown
	return mgr
}

// SetStartupFunc can be used to optionally specify a startup function that
// will be called before calling run.
func (m *Manager) SetStartupFunc(fn func(context.Context) error) error {
	s := <-m.status
	switch s {
	case StatusShutdown:
		m.status <- s
		return ErrShutdown
	case StatusUnknown:
		m.startupFunc = fn
		m.status <- s
		return nil
	default:
		m.status <- s
		return ErrAlreadyStarted
	}
}

// SetPauseResumer will set the PauseResumer used by Pause and Resume methods.
func (m *Manager) SetPauseResumer(pr PauseResumer) error {
	s := <-m.status
	if m.isPausing || s == StatusPausing || s == StatusPaused {
		m.status <- s
		return errors.New("cannot SetPauseResumer during pause operation")
	}
	m.pauseResume = pr
	m.status <- s
	return nil
}

// IsPausing will return true if the manager is in a state of
// pause, or is currently fulfilling a Pause request.
func (m *Manager) IsPausing() bool {
	s := <-m.status
	isPausing := m.isPausing
	m.status <- s
	switch s {
	case StatusPausing, StatusPaused:
		return true
	case StatusShutdown:
		return true
	}
	return isPausing
}

// PauseWait will return a channel that blocks until a pause operation begins.
func (m *Manager) PauseWait() <-chan struct{} {
	s := <-m.status
	ch := m.pauseStart
	m.status <- s
	return ch
}

// WaitForStartup will wait for startup to complete (even if failed or shutdown).
// err is nil unless context deadline is reached or startup produced an error.
func (m *Manager) WaitForStartup(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-m.startupDone:
		return m.startupErr
	}
}

// Status returns the current status.
func (m *Manager) Status() Status {
	s := <-m.status
	m.status <- s
	return s
}

// Run starts the main loop.
func (m *Manager) Run(ctx context.Context) error {
	s := <-m.status
	switch s {
	case StatusShutdown:
		m.status <- s
		return ErrShutdown
	case StatusUnknown:
		// ok
	default:
		m.status <- s
		return ErrAlreadyStarted
	}

	startCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	m.startupCancel = cancel
	startupFunc := m.startupFunc
	m.status <- StatusStarting

	if startupFunc != nil {
		m.startupErr = startupFunc(startCtx)
	}
	cancel()

	s = <-m.status

	switch s {
	case StatusShutdown:
		m.status <- s
		// no error on shutdown while starting
		return nil
	case StatusStarting:
		if m.startupErr != nil {
			m.status <- s
			close(m.startupDone)
			return m.startupErr
		}
		// ok
	default:
		m.status <- s
		panic("unexpected lifecycle state")
	}

	ctx, m.runCancel = context.WithCancel(ctx)
	close(m.startupDone)
	m.status <- StatusReady

	err := m.runFunc(ctx)
	close(m.runDone)
	s = <-m.status
	m.status <- s
	if s == StatusShutdown {
		<-m.shutdownDone
	}

	return err
}

// Shutdown begins the shutdown procedure.
func (m *Manager) Shutdown(ctx context.Context) error {
	initShutdown := func() {
		ctx, m.shutdownCancel = context.WithCancel(ctx)
		m.status <- StatusShutdown
	}

	var isRunning bool
	s := <-m.status
	switch s {
	case StatusShutdown:
		m.status <- s
		select {
		case <-m.shutdownDone:
		case <-ctx.Done():
			// if we timeout before the existing call, cancel it's context
			m.shutdownCancel()
			<-m.shutdownDone
		}
		return m.shutdownErr
	case StatusStarting:
		m.startupCancel()
		close(m.pauseStart)
		initShutdown()
		<-m.startupDone
	case StatusUnknown:
		initShutdown()
		close(m.pauseStart)
		close(m.shutdownDone)
		return nil
	case StatusPausing:
		isRunning = true
		m.pauseCancel()
		initShutdown()
		<-m.pauseDone
	case StatusReady:
		close(m.pauseStart)
		fallthrough
	case StatusPaused:
		isRunning = true
		initShutdown()
	}

	defer close(m.shutdownDone)
	defer m.shutdownCancel()

	err := m.shutdownFunc(ctx)

	if isRunning {
		m.runCancel()
		<-m.runDone
	}

	return err
}

// Pause will bein a pause operation.
// SetPauseResumer must have been called or ErrPauseUnsupported is returned.
//
// Pause is atomic and guarantees a paused state if nil is returned
// or normal operation otherwise.
func (m *Manager) Pause(ctx context.Context) error {
	s := <-m.status
	if m.pauseResume == nil {
		m.status <- s
		return ErrPauseUnsupported
	}
	switch s {
	case StatusShutdown:
		m.status <- s
		return ErrShutdown
	case StatusPaused:
		m.status <- s
		return nil
	case StatusPausing:
		pauseDone := m.pauseDone
		m.status <- s
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pauseDone:
			return m.Pause(ctx)
		}
	case StatusStarting, StatusUnknown:
		if m.isPausing {
			pauseDone := m.pauseDone
			m.status <- s
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-pauseDone:
				return m.Pause(ctx)
			}
		}
	case StatusReady:
		// ok
	}

	ctx, m.pauseCancel = context.WithCancel(ctx)
	m.pauseDone = make(chan struct{})
	m.isPausing = true
	defer close(m.pauseDone)
	defer m.pauseCancel()
	m.pauseErr = nil
	if s != StatusReady {
		m.status <- s
		select {
		case <-ctx.Done():
			s = <-m.status
			m.isPausing = false
			m.status <- s
			return ctx.Err()
		case <-m.startupDone:
		}

		s = <-m.status
		switch s {
		case StatusShutdown:
			m.status <- s
			return ErrShutdown
		case StatusReady:
			// ok
		default:
			m.status <- s
			panic("unexpected lifecycle state")
		}
	}

	close(m.pauseStart)
	m.status <- StatusPausing
	err := m.pauseResume.Pause(ctx)
	m.pauseCancel()
	s = <-m.status
	switch s {
	case StatusShutdown:
		m.pauseErr = ErrShutdown
		m.isPausing = false
		m.status <- s
		return ErrShutdown
	case StatusPausing:
		// ok
	default:
		m.isPausing = false
		m.status <- s
		panic("unexpected lifecycle state")
	}

	if err != nil {
		m.pauseErr = err
		m.isPausing = false
		m.pauseStart = make(chan struct{})
		m.status <- StatusReady
		return err
	}

	m.pauseErr = nil
	m.isPausing = false
	m.status <- StatusPaused
	return nil
}

// Resume will always result in normal operation (unless Shutdown was called).
//
// If the context deadline is reached, "graceful" operations may fail, but
// will always result in a Ready state.
func (m *Manager) Resume(ctx context.Context) error {
	s := <-m.status
	if m.pauseResume == nil {
		m.status <- s
		return ErrPauseUnsupported
	}
	switch s {
	case StatusShutdown:
		m.status <- s
		return ErrShutdown
	case StatusUnknown, StatusStarting:
		if !m.isPausing {
			m.status <- s
			return nil
		}

		fallthrough
	case StatusPausing:
		m.pauseCancel()
		pauseDone := m.pauseDone
		m.status <- s
		<-pauseDone
		return m.Resume(ctx)
	case StatusPaused:
		// ok
	case StatusReady:
		m.status <- s
		return nil
	}

	m.pauseStart = make(chan struct{})
	err := m.pauseResume.Resume(ctx)
	m.status <- StatusReady

	return err
}
