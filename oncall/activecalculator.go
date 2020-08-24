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

func (t *TimeIterator) NewActiveCalculator() *ActiveCalculator {
	act := &ActiveCalculator{
		TimeIterator: t,
		states:       boolMapPool.Get().([]boolValue),
		times:        timeMapPool.Get().([]time.Time),
	}
	t.Register(act.next, act.done)

	return act
}

func (act *ActiveCalculator) Init() *ActiveCalculator {
	if act.init {
		return act
	}
	act.init = true

	sort.Sort((*activeSortable)(act))

	return act
}

func (act *ActiveCalculator) SetSpan(start, end time.Time) {
	if act.init {
		panic("cannot add spans after Init")
	}
	if !end.After(act.Start()) && !end.IsZero() {
		return
	}
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

func (act *ActiveCalculator) next(t int64) int64 {
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
func (act *ActiveCalculator) done() {
	//lint:ignore SA6002 not worth the overhead to avoid the slice-struct allocation
	boolMapPool.Put(act.states[:0])
	//lint:ignore SA6002 not worth the overhead to avoid the slice-struct allocation
	timeMapPool.Put(act.times[:0])
	act.states, act.times = nil, nil
}
func (act *ActiveCalculator) Active() bool          { return act.active }
func (act *ActiveCalculator) Changed() bool         { return act.changed }
func (act *ActiveCalculator) ActiveTime() time.Time { return act.activeT }
