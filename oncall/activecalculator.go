package oncall

import (
	"time"
)

type ActiveCalculator struct {
	*TimeIterator
	m map[int64]bool

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

func (act *ActiveCalculator) SetSpan(start, end time.Time) {
	if !end.After(act.Start()) {
		return
	}
	if !start.Before(act.End()) {
		return
	}
	if start.Before(act.Start()) {
		start = act.Start()
	}

	act.set(start, true)
	act.set(end, false)
}
func (act *ActiveCalculator) set(t time.Time, isStart bool) {
	act.m[t.Truncate(act.Step()).Unix()] = isStart
}

func (act *ActiveCalculator) next() {
	val, ok := act.m[act.Unix()]
	act.changed = ok
	if ok {
		act.active = val
	}
}
func (act *ActiveCalculator) Active() bool  { return act.active }
func (act *ActiveCalculator) Changed() bool { return act.changed }
