package app

import (
	"context"
	"time"
)

type cooldown struct {
	dur    time.Duration
	trigCh chan struct{}
	waitCh chan struct{}
}

func newCooldown(dur time.Duration) *cooldown {
	c := &cooldown{
		dur:    dur,
		trigCh: make(chan struct{}),
		waitCh: make(chan struct{}),
	}
	go c.loop()
	return c
}
func (c *cooldown) loop() {
	t := time.NewTimer(c.dur)
	var active bool

	for {
		if active {
			select {
			case <-t.C:
				active = false
			case c.trigCh <- struct{}{}:
				t.Stop()
				t = time.NewTimer(c.dur)

				active = true
			}
			continue
		}

		select {
		// not active, allow closing
		case c.waitCh <- struct{}{}:
		case c.trigCh <- struct{}{}:
			t.Stop()
			t = time.NewTimer(c.dur)
			active = true
		}
	}
}

// WaitContext will wait until there have been no new connections within the cooldown period.
func (c *cooldown) WaitContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.waitCh:
		return nil
	}
}

func (c *cooldown) Trigger() {
	<-c.trigCh
}
