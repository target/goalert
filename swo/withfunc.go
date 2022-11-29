package swo

import (
	"context"
	"fmt"
	"sync"
)

// WithFunc flattens a with-type func providing it's value with Begin() and ending with Cancel().
type WithFunc[V any] struct {
	withFn func(context.Context, func(V)) error

	mx     sync.Mutex
	wg     sync.WaitGroup
	cancel func()
}

// NewWithFunc creates a new WithFunc.
//
// withFn must not return an error after useFn is called.
func NewWithFunc[V any](withFn func(ctx context.Context, useFn func(V)) error) *WithFunc[V] {
	return &WithFunc[V]{
		withFn: withFn,
	}
}

// Begin will return a new instance of V, Cancel should be called when it's no longer needed.
//
// If err is nil, Cancel must be called before calling Begin again.
func (w *WithFunc[V]) Begin(ctx context.Context) (v V, err error) {
	w.mx.Lock()
	defer w.mx.Unlock()

	if w.cancel != nil {
		return v, fmt.Errorf("already in progress")
	}

	ctx, w.cancel = context.WithCancel(ctx)

	ch := make(chan V, 1)
	errCh := make(chan error, 1)
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()
		var called bool
		err := w.withFn(ctx, func(v V) {
			called = true
			ch <- v
			<-ctx.Done()
		})
		if err == nil {
			if !called {
				errCh <- fmt.Errorf("useFn never called")
			}
			return
		}

		if called {
			panic(fmt.Errorf("error returned after withFn called: %w", err))
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		w._cancel()
		return v, ctx.Err()
	case err = <-errCh:
		w._cancel()
		return v, err
	case v = <-ch:
		return v, nil
	}
}

// Cancel will cancel context passed to withFn and wait for it to finish.
func (w *WithFunc[V]) Cancel() {
	w.mx.Lock()
	defer w.mx.Unlock()

	if w.cancel == nil {
		return
	}

	w._cancel()
}

func (w *WithFunc[V]) _cancel() {
	w.cancel()
	w.cancel = nil

	w.wg.Wait()
}
