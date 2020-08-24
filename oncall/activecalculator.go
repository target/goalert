package oncall

import (
	"sync"
	"time"
)

var (
	boolMapPool = &sync.Pool{
		New: func() interface{} { return make(map[int64]bool, 20) },
	}
	timeMapPool = &sync.Pool{
		New: func() interface{} { return make(map[int64]time.Time, 20) },
	}
)

type ActiveCalculator struct {
	*TimeIterator
	m        map[int64]bool
	actStart map[int64]time.Time

	activeT time.Time
	active  bool
	changed bool
}

func (t *TimeIterator) NewActiveCalculator() *ActiveCalculator {
	act := &ActiveCalculator{
		TimeIterator: t,
		m:            boolMapPool.Get().(map[int64]bool),
		actStart:     timeMapPool.Get().(map[int64]time.Time),
	}
	t.Register(act.next, act.done)

	return act
}

func (act *ActiveCalculator) SetSpan(start, end time.Time) {
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
	if isStart {
		act.actStart[id] = t.Truncate(act.Step())
	}
	act.m[id] = isStart
}

func (act *ActiveCalculator) next() {
	val, ok := act.m[act.Unix()]
	act.changed = ok
	if ok {
		act.active = val
		act.activeT = act.actStart[act.Unix()]
	}
}
func (act *ActiveCalculator) done() {
	for id := range act.m {
		delete(act.m, id)
	}
	for id := range act.actStart {
		delete(act.actStart, id)
	}
	boolMapPool.Put(act.m)
	timeMapPool.Put(act.actStart)
	act.m, act.actStart = nil, nil
}
func (act *ActiveCalculator) Active() bool          { return act.active }
func (act *ActiveCalculator) Changed() bool         { return act.changed }
func (act *ActiveCalculator) ActiveTime() time.Time { return act.activeT }
