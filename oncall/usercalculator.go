package oncall

import "time"

type UserCalculator struct {
	*TimeIterator

	m map[string]*ActiveCalculator

	calc []userCalc

	init    bool
	active  []string
	changed bool
}
type userCalc struct {
	ID   string
	Calc *ActiveCalculator
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

	u.calc = make([]userCalc, 0, len(u.m))
	for id, a := range u.m {
		a.Init()
		u.calc = append(u.calc, userCalc{ID: id, Calc: a})
	}

	u.Register(u.next, nil)
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
func (u *UserCalculator) next(int64) int64 {
	if !u.init {
		panic("init was never called")
	}
	if len(u.calc) == 0 {
		return -1
	}
	u.changed = false
	for _, c := range u.calc {
		if !c.Calc.Changed() {
			continue
		}
		u.changed = true
		if c.Calc.Active() {
			u.active = append(u.active, c.ID)
		} else {
			u.active = omitStr(u.active, c.ID)
		}
	}

	return 0
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
