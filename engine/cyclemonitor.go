package engine

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/target/goalert/validation"
)

const cycleHist = 10

type cycleMonitor struct {
	mx sync.Mutex

	cycles  map[uuid.UUID]chan struct{}
	history [cycleHist]uuid.UUID
}

func newCycleMonitor() *cycleMonitor {
	m := &cycleMonitor{
		cycles: make(map[uuid.UUID]chan struct{}, cycleHist),
	}
	m._newID()
	return m
}

func (c *cycleMonitor) _newID() {
	// remove oldest cycle
	delete(c.cycles, c.history[cycleHist-1])

	// shift history
	copy(c.history[:], c.history[1:])

	// add new cycle
	c.history[0] = uuid.New()
	c.cycles[c.history[0]] = make(chan struct{})
}

// startNextCycle marks the beginning of the next engine cycle.
// It returns a func that should be called when the cycle is finished.
func (c *cycleMonitor) startNextCycle() func() {
	c.mx.Lock()
	defer c.mx.Unlock()

	ch := c.cycles[c.history[0]]
	c._newID()

	return func() { close(ch) }
}

// WaitCycleID waits for the engine cycle with the given UUID to finish.
func (c *cycleMonitor) WaitCycleID(ctx context.Context, cycleID uuid.UUID) error {
	if c == nil {
		// engine is disabled
		return validation.NewGenericError("engine is disabled")
	}

	c.mx.Lock()
	ch, ok := c.cycles[cycleID]
	c.mx.Unlock()
	if !ok {
		return validation.NewGenericError("unknown cycle ID")
	}

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// NextCycleID returns the UUID of the next engine cycle.
func (c *cycleMonitor) NextCycleID() uuid.UUID {
	if c == nil {
		// engine is disabled
		return uuid.Nil
	}

	c.mx.Lock()
	defer c.mx.Unlock()

	return c.history[0]
}
