package oncall

import (
	"time"
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
		m:            make(map[int64]bool),
		actStart:     make(map[int64]time.Time),
	}
	t.OnNext(act.next)

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
	act.set(end, false)
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
func (act *ActiveCalculator) Active() bool          { return act.active }
func (act *ActiveCalculator) Changed() bool         { return act.changed }
func (act *ActiveCalculator) ActiveTime() time.Time { return act.activeT }
