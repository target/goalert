package oncall

import (
	"sort"
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
	ID         int64
	Value      bool
	OriginalID int64
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

// Init should be called after all SetSpan calls have been completed and before Next().
func (act *ActiveCalculator) Init() *ActiveCalculator {
	if act.init {
		return act
	}
	act.init = true

	sort.Slice(act.states, func(i, j int) bool { return act.states[i].ID < act.states[j].ID })

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

	// Skip if the span ends before the iterator start time.
	//
	// A zero end time indicates infinity (e.g. current shift from history).
	if !end.After(act.Start()) && !end.IsZero() {
		return
	}

	// Skip if the length of the span is <= 0.
	if !start.Before(act.End()) {
		return
	}

	act.set(start, true)
	if !end.IsZero() {
		act.set(end, false)
	}
}

func (act *ActiveCalculator) set(t time.Time, isStart bool) {
	id := t.Truncate(act.Step()).Unix()
	originalID := id
	if isStart && t.Before(act.Start()) {
		id = act.Start().Unix()
	}

	if len(act.states) > 0 && isStart && id == act.states[len(act.states)-1].ID {
		act.states = act.states[:len(act.states)-1]
		return
	}

	act.states = append(act.states, activeCalcValue{ID: id, Value: isStart, OriginalID: originalID})
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

	v := act.states[0]
	act.changed = v.ID == t
	if act.changed {
		act.active = v
		act.states = act.states[1:]
		if len(act.states) > 0 {
			// if act.states[0].ID <= t {
			// 	panic("WAS LESS")
			// }
			return act.states[0].ID
		}

		return -1
	}

	return v.ID
}

// Done implements the SubIterator.Done method.
func (act *ActiveCalculator) Done() {
	//lint:ignore SA6002 not worth the overhead to avoid the slice-struct allocation
	activeCalcValuePool.Put(act.states[:0])

	act.states = nil
}

// Active will return true if the current timestamp is within a span.
func (act *ActiveCalculator) Active() bool { return act.active.Value }

// Changed will return true if the current tick changed the Active() state.
func (act *ActiveCalculator) Changed() bool { return act.changed }

// ActiveTime returns the original start time of the current Active() state.
//
// If Active() is false, it returns a zero value.
func (act *ActiveCalculator) ActiveTime() time.Time {
	if !act.Active() {
		return time.Time{}
	}

	return time.Unix(act.active.OriginalID, 0).UTC()
}
