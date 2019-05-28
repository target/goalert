package util

import (
	"math/rand"
	"sync"
	"time"
)

// AlignedTicker works like a time.Ticker except it will align the first tick.
// This makes it useful in situations where something should run on-the-minute
// for example.
type AlignedTicker struct {
	tm   *time.Timer
	tc   *time.Ticker
	mx   sync.Mutex
	done bool
	dur  time.Duration
	c    chan time.Time
	C    <-chan time.Time
}

// NewAlignedTicker will create and start a new AlignedTicker. The first tick
// will be adjusted to round, with variance added.
//
// For example (time.Minute, time.Second) will align ticks to on-the-minute
// plus 0-1 second.
func NewAlignedTicker(round, variance time.Duration) *AlignedTicker {
	vary := time.Duration(rand.Int63n(int64(variance)))
	s := time.Now().Round(round).Add(vary)
	for s.Before(time.Now()) {
		s = s.Add(round)
	}
	a := &AlignedTicker{
		c:   make(chan time.Time),
		dur: round,
	}
	a.C = a.c
	a.tm = time.AfterFunc(time.Until(s), a.firstTick)
	return a
}
func (a *AlignedTicker) firstTick() {
	a.mx.Lock()
	defer a.mx.Unlock()
	if a.done {
		return
	}

	a.tc = time.NewTicker(a.dur)
	go func() {
		for t := range a.tc.C {
			a.mx.Lock()
			if a.done {
				a.mx.Unlock()
				break
			}
			a.c <- t
			a.mx.Unlock()
		}
	}()
}

// Stop will stop the running ticker and close the channel.
func (a *AlignedTicker) Stop() {
	a.mx.Lock()
	defer a.mx.Unlock()
	if a.done {
		return
	}
	a.done = true
	a.tm.Stop()
	if a.tc != nil {
		a.tc.Stop()
	}
	close(a.c)
}
