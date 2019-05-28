package lifecycle

import (
	"context"

	"github.com/pkg/errors"
)

// A PauseResumer can be atomically paused and resumed.
type PauseResumer interface {
	// Pause should result in pausing all operations if nil is returned.
	//
	// If a pause cannot complete within the context deadline,
	// the context error should be returned, and normal operation should
	// resume, as if pause was never called.
	Pause(context.Context) error

	// Resume should always result in normal operation.
	//
	// Context can be used for control of graceful operations,
	// but Resume should not return until normal operation is restored.
	//
	// Operations that are required for resuming, should use a background context
	// internally (possibly linking any trace spans).
	Resume(context.Context) error
}

type prFunc struct{ pause, resume func(context.Context) error }

func (p prFunc) Pause(ctx context.Context) error  { return p.pause(ctx) }
func (p prFunc) Resume(ctx context.Context) error { return p.resume(ctx) }

var _ PauseResumer = prFunc{}

// PauseResumerFunc is a convenience method that takes a pause and resume func
// and returns a PauseResumer.
func PauseResumerFunc(pause, resume func(context.Context) error) PauseResumer {
	return prFunc{pause: pause, resume: resume}
}

// MultiPauseResume will join multiple PauseResumers where
// all will be paused, or none.
//
// Any that pause successfully, when another fails, will
// have Resume called.
func MultiPauseResume(pr ...PauseResumer) PauseResumer {
	pause := func(ctx context.Context) error {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		pass := make(chan struct{})
		fail := make(chan struct{})
		errCh := make(chan error, len(pr))
		resumeErrCh := make(chan error, len(pr))

		doPause := func(p PauseResumer) {
			err := errors.Wrapf(p.Pause(ctx), "pause")
			errCh <- err
			select {
			case <-pass:
				resumeErrCh <- nil
			case <-fail:
				if err == nil {
					resumeErrCh <- errors.Wrapf(p.Resume(ctx), "resume")
				} else {
					resumeErrCh <- nil
				}
			}
		}

		for _, p := range pr {
			go doPause(p)
		}

		var hasErr bool
		var errs []error
		for range pr {
			err := <-errCh
			if err != nil {
				errs = append(errs, err)
				if !hasErr {
					cancel()
					close(fail)
					hasErr = true
				}
			}
		}
		if !hasErr {
			close(pass)
		}
		for range pr {
			err := <-resumeErrCh
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errors.Errorf("multiple errors: %v", errs)
		}

		return nil
	}
	resume := func(ctx context.Context) error {
		ch := make(chan error)
		res := func(fn func(context.Context) error) { ch <- fn(ctx) }
		for _, p := range pr {
			go res(p.Resume)
		}
		var errs []error
		for range pr {
			err := <-ch
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) > 0 {
			return errors.Errorf("multiple errors: %v", errs)
		}
		return nil
	}

	return PauseResumerFunc(pause, resume)
}
