package oncall

import (
	"time"
)

type ActiveCalculator struct {
	*TimeIterator
	m map[int64]bool

	closestToStart time.Time
	startActive    bool

	init    bool
	active  bool
	changed bool
}

func (t *TimeIterator) NewActiveCalculator() *ActiveCalculator {
	act := &ActiveCalculator{
		TimeIterator: t,
		m:            make(map[int64]bool),
	}
	t.OnNext(act.next)

	return act
}
func (act *ActiveCalculator) Init() *ActiveCalculator {
	if act.init {
		return act
	}
	act.init = true

	_, ok := act.m[act.start]
	if ok {
		return act
	}
	if act.startActive {
		act.m[act.start] = true
	}

	return act
}
func (act *ActiveCalculator) SetSpan(start, end time.Time) {
	if act.init {
		panic("cannot call SetSpan after Init")
	}
	act.set(start, true)
	act.set(end, false)
}
func (act *ActiveCalculator) set(t time.Time, isStart bool) {
	if t.Before(act.Start()) && t.After(act.closestToStart) {
		act.closestToStart = t
		act.startActive = isStart
	}

	// skip out of bounds
	if !isStart && !t.After(act.Start()) {
		return
	}
	if isStart && !t.Before(act.End()) {
		return
	}

	act.m[t.Truncate(act.Step()).Unix()] = isStart
}

func (act *ActiveCalculator) next() {
	if !act.init {
		panic("Init never called")
	}
	val, ok := act.m[act.Unix()]
	act.changed = ok
	if ok {
		act.active = val
	}
}
func (act *ActiveCalculator) Active() bool  { return act.active }
func (act *ActiveCalculator) Changed() bool { return act.changed }
