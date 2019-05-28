package slack

import (
	"context"
	"sync"
	"time"
)

type throttle struct {
	tick      *time.Ticker
	waitUntil time.Time
	mx        sync.Mutex
}

func newThrottle(dur time.Duration) *throttle {
	return &throttle{
		tick: time.NewTicker(dur),
	}
}

func (t *throttle) Wait(ctx context.Context) error {
	t.mx.Lock()
	dur := time.Until(t.waitUntil)
	t.mx.Unlock()

	if dur > 0 {
		tm := time.NewTimer(dur)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tm.C:
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.tick.C:
	}

	return nil
}

func (t *throttle) SetWaitUntil(end time.Time) {
	t.mx.Lock()
	if end.After(t.waitUntil) {
		t.waitUntil = end
	}
	t.mx.Unlock()
}
