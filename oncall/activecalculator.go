package oncall

import (
	"fmt"
	"sync"
	"time"
)

var (
	activeCalcValuePool = &sync.Pool{
		New: func() interface{} { return make([]activeCalcValue, 0, 100) },
	}
)

// ActiveCalculator will calculate if the current timestamp is within a span.
type ActiveCalculator struct {
	*TimeIterator

	states []activeCalcValue

	init    bool
	active  activeCalcValue
	changed bool
}
type activeCalcValue struct {
	T       int64
	IsStart bool
}

// NewActiveCalculator will create a new ActiveCalculator bound to the TimeIterator.
func (t *TimeIterator) NewActiveCalculator() *ActiveCalculator {
	act := &ActiveCalculator{
		TimeIterator: t,
		states:       activeCalcValuePool.Get().([]activeCalcValue),
	}
	t.Register(act)

	return act
}

func (act *ActiveCalculator) StartUnix() (start int64) {
	if len(act.states) == 0 {
		return 0
	}

	return act.states[0].T
}

// Init should be called after all SetSpan calls have been completed and before Next().
func (act *ActiveCalculator) Init() *ActiveCalculator {
	if act.init {
		return act
	}
	act.init = true

	return act
}

// SetSpan is used to set an active span.
//
// Care should be taken so that there is no overlap between spans, and
// no start time should equal any end time for non-sequential calls.
func (act *ActiveCalculator) SetSpan(start, end time.Time) {
	if act.init {
		panic("cannot add spans after Init")
	}
	start = start.Truncate(act.Step())
	end = end.Truncate(act.Step())

	// Skip if the span is < 1 step (after truncation).
	if !end.After(start) {
		return
	}

	// Skip if the span ends before or at the iterator start time.
	if !end.After(act.Start()) {
		return
	}

	// Skip if the span starts at or after the calculator end time.
	if !start.Before(act.End()) {
		return
	}

	act.set(start, true)
	act.set(end, false)
}

func (act *ActiveCalculator) set(t time.Time, isStart bool) {
	id := t.Unix()

	if len(act.states) == 0 && !isStart {
		panic("end registered before start")
	}
	if len(act.states) > 0 && isStart == act.states[len(act.states)-1].IsStart {
		panic("must not overlap shifts")
	}

	if len(act.states) > 0 && isStart && id == act.states[len(act.states)-1].T {
		// starting a shift at the same time one ends, so just delete the end marker
		act.states = act.states[:len(act.states)-1]
		return
	}

	if len(act.states) > 0 && id <= act.states[len(act.states)-1].T {
		panic(fmt.Sprintf("shifts must be registered in order: got %d, want > %d in %#v", id, act.states[len(act.states)-1].T, act.states))
	}

	act.states = append(act.states, activeCalcValue{T: id, IsStart: isStart})
}

// Process implements the SubIterator.Process method.
func (act *ActiveCalculator) Process(t int64) int64 {
	if !act.init {
		panic("Init never called")
	}
	if len(act.states) == 0 {
		act.changed = false
		return -1
	}

	val := act.states[0]
	act.changed = val.T == t
	if act.changed {
		act.active = val
		act.states = act.states[1:]
		if len(act.states) > 0 {
			return act.states[0].T
		}

		return -1
	}

	return val.T
}

// Done implements the SubIterator.Done method.
func (act *ActiveCalculator) Done() {
	//lint:ignore SA6002 not worth the overhead to avoid the slice-struct allocation
	activeCalcValuePool.Put(act.states[:0])

	act.states = nil
}

// Active will return true if the current timestamp is within a span.
func (act *ActiveCalculator) Active() bool { return act.active.IsStart }

// Changed will return true if the current tick changed the Active() state.
func (act *ActiveCalculator) Changed() bool { return act.changed }
