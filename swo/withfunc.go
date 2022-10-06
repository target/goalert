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

func NewWithFunc[V any](withFn func(context.Context, func(V)) error) *WithFunc[V] {
	return &WithFunc[V]{
		withFn: withFn,
	}
}

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
		errCh <- w.withFn(ctx, func(v V) {
			ch <- v
			<-ctx.Done()
		})
	}()

	select {
	case <-ctx.Done():
		return v, ctx.Err()
	case err = <-errCh:
		return v, err
	case v = <-ch:
		return v, nil
	}
}

func (w *WithFunc[V]) Cancel() {
	w.mx.Lock()
	defer w.mx.Unlock()

	if w.cancel == nil {
		return
	}

	w.cancel()
	w.cancel = nil

	w.wg.Wait()
}
