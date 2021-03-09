package oncall

import "time"

// UserCalculator will calculate a set of active users for a set of time spans.
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

// NewUserCalculator will create a new UserCalculator bound to the TimeIterator.
func (t *TimeIterator) NewUserCalculator() *UserCalculator {
	u := &UserCalculator{
		TimeIterator: t,
		m:            make(map[string]*ActiveCalculator),
	}

	return u
}

// Init should be called after all SetSpan values have been provided.
func (u *UserCalculator) Init() *UserCalculator {
	if u.init {
		return u
	}

	// transition to slice to avoid map iteration in Process
	u.calc = make([]userCalc, 0, len(u.m))
	for id, a := range u.m {
		a.Init()
		u.calc = append(u.calc, userCalc{ID: id, Calc: a})
	}

	u.Register(u)
	u.init = true

	return u
}

// SetSpan is used to set a start & end time for the given user ID.
//
// Care should be taken so that there is no overlap between spans of the same id, and
// no start time should equal any end time for the same id.
func (u *UserCalculator) SetSpan(start, end time.Time, id string) {
	if u.init {
		panic("cannot call SetSpan after init")
	}

	// set span per UserID
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

// Process implements the SubIterator.Process method.
func (u *UserCalculator) Process(int64) int64 {
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

// Done implements the SubIterator.Done method.
func (u *UserCalculator) Done() {}

// ActiveUsers returns the current set of active users for the current timestamp.
//
// It is only valid until the following Next() call and should not be modified.
func (u *UserCalculator) ActiveUsers() []string { return u.active }

// Changed will return true if there has been any change this tick.
func (u *UserCalculator) Changed() bool { return u.changed }

// ActiveTimes will return the original start time for all ActiveUsers.
func (u *UserCalculator) ActiveTimes() []time.Time {
	times := make([]time.Time, len(u.active))
	for i, id := range u.active {
		times[i] = u.m[id].ActiveTime()
	}
	return times
}
