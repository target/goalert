package oncall

import "time"

type UserCalculator struct {
	*TimeIterator

	m map[string]*ActiveCalculator

	init    bool
	active  []string
	start   []time.Time
	changed bool
}

func (t *TimeIterator) NewUserCalculator() *UserCalculator {
	u := &UserCalculator{
		TimeIterator: t,
		m:            make(map[string]*ActiveCalculator),
	}

	return u
}
func (u *UserCalculator) Init() *UserCalculator {
	if u.init {
		return u
	}

	u.OnNext(u.next)
	u.init = true

	return u
}

func (u *UserCalculator) SetSpan(start, end time.Time, id string) {
	if u.init {
		panic("cannot call SetSpan after init")
	}
	c := u.m[id]
	if c == nil {
		c = u.NewActiveCalculator()
		u.m[id] = c
	}

	c.SetSpan(start, end)
}
func omitStr(s []string, val string) []string {
	idx := -1
	for i, str := range s {
		if str != val {
			continue
		}
		idx = i
		break
	}
	if idx == -1 {
		return s
	}
	return append(s[:idx], s[idx+1:]...)
}
func (u *UserCalculator) next() {
	if !u.init {
		panic("init was never called")
	}
	u.changed = false
	for id, calc := range u.m {
		if !calc.Changed() {
			continue
		}
		u.changed = true
		if calc.Active() {
			u.active = append(u.active, id)
		} else {
			u.active = omitStr(u.active, id)
		}
	}
}

func (u *UserCalculator) ActiveUsers() []string { return u.active }
func (u *UserCalculator) Changed() bool         { return u.changed }
func (u *UserCalculator) ActiveTimes() []time.Time {
	times := make([]time.Time, len(u.active))
	for i, id := range u.active {
		times[i] = u.m[id].ActiveTime()
	}
	return times
}
