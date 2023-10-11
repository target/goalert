package ctxlock

import (
	"context"
	"time"
)

// Lock will attempt to acquire a lock for the given ID.
//
// If Lock returns nil, Unlock must be called to release the lock,
// even if the context is canceled.
func (l *IDLocker[K]) Lock(ctx context.Context, id K) error {
	l.mx.Lock()
	if l.count[id] < l.cfg.MaxHeld {
		// fast path, no queue
		l.count[id]++
		l.mx.Unlock()
		return nil
	}

	if l.cfg.MaxWait != -1 && len(l.queue[id]) >= l.cfg.MaxWait {
		l.mx.Unlock()
		return ErrQueueFull
	}

	if l.cfg.Timeout > 0 {
		var cancel func(error)
		ctx, cancel = context.WithCancelCause(ctx)
		t := time.AfterFunc(l.cfg.Timeout, func() { cancel(ErrTimeout) })
		defer t.Stop()
		defer cancel(nil)
	}

	// slow path, queue to hold our spot
	ch := make(chan struct{})
	l.queue[id] = append(l.queue[id], ch)
	l.mx.Unlock()

	select {
	case <-ctx.Done():
		close(ch) // Ensure Unlock knows we are abandoning our spot.

		l.mx.Lock()
		defer l.mx.Unlock()
		for i, c := range l.queue[id] {
			if c != ch {
				continue
			}

			l.queue[id] = append(l.queue[id][:i], l.queue[id][i+1:]...)
			break
		}
		if len(l.queue[id]) == 0 {
			delete(l.queue, id) // cleanup so the map doesn't grow forever
		}

		return context.Cause(ctx)
	case ch <- struct{}{}:
		// we have the lock, queue and count have been updated by Unlock
	}

	return nil
}
