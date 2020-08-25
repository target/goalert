package oncall

import (
	"sort"
	"sync"
	"time"
)

var (
	boolMapPool = &sync.Pool{
		New: func() interface{} { return make([]boolValue, 0, 100) },
	}
	timeMapPool = &sync.Pool{
		New: func() interface{} { return make([]time.Time, 0, 100) },
	}
)

// ActiveCalculator will calculate if the current timestamp is within a span.
type ActiveCalculator struct {
	*TimeIterator

	states []boolValue
	times  []time.Time

	init    bool
	activeT time.Time
	active  bool
	changed bool
}
type boolValue struct {
	ID    int64
	Value bool
}

type activeSortable ActiveCalculator

func (act *activeSortable) Less(i, j int) bool {
	return act.states[i].ID < act.states[j].ID
}
func (act *activeSortable) Len() int { return len(act.states) }
func (act *activeSortable) Swap(i, j int) {
	act.states[i], act.states[j] = act.states[j], act.states[i]
	act.times[i], act.times[j] = act.times[j], act.times[i]
}

// NewActiveCalculator will create a new ActiveCalculator bound to the TimeIterator.
func (t *TimeIterator) NewActiveCalculator() *ActiveCalculator {
	act := &ActiveCalculator{
		TimeIterator: t,
		states:       boolMapPool.Get().([]boolValue),
		times:        timeMapPool.Get().([]time.Time),
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

	sort.Sort((*activeSortable)(act))

	return act
}

// SetSpan is used to set an active span.
//
// Care should be taken so that there is no overlap between spans, and
// no start time should equal any end time.
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
	if isStart && t.Before(act.Start()) {
		id = act.Start().Unix()
	}

	act.times = append(act.times, t.Truncate(act.Step()))
	act.states = append(act.states, boolValue{ID: id, Value: isStart})
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
		act.states = act.states[1:]
		act.active = v.Value
		act.activeT = act.times[0]
		act.times = act.times[1:]
		if len(act.states) > 0 {
			return act.states[0].ID
		}

		return -1
	}

	return v.ID
}

// Done implements the SubIterator.Done method.
func (act *ActiveCalculator) Done() {
	//lint:ignore SA6002 not worth the overhead to avoid the slice-struct allocation
	boolMapPool.Put(act.states[:0])
	//lint:ignore SA6002 not worth the overhead to avoid the slice-struct allocation
	timeMapPool.Put(act.times[:0])
	act.states, act.times = nil, nil
}

// Active will return true if the current timestamp is within a span.
func (act *ActiveCalculator) Active() bool { return act.active }

// Changed will return true if the current tick changed the Active() state.
func (act *ActiveCalculator) Changed() bool { return act.changed }

// ActiveTime returns the original start time of the current Active() state.
//
// It is only valid if Active() is true.
func (act *ActiveCalculator) ActiveTime() time.Time { return act.activeT }
