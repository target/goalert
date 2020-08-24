package oncall

import "time"

type TimeIterator struct {
	t, start, end, step int64

	next []func()
	done []func()
}

type Iterator interface {
	Next()
	Done()
}

func NewTimeIterator(start, end time.Time, step time.Duration) *TimeIterator {
	step = step.Truncate(time.Second)
	stepUnix := step.Nanoseconds() / int64(time.Second)
	start = start.Truncate(step)
	end = end.Truncate(step)

	return &TimeIterator{
		step:  stepUnix,
		start: start.Unix(),
		end:   end.Unix(),
		t:     start.Unix() - stepUnix,
	}
}

func (iter *TimeIterator) Register(next, done func()) {
	if next != nil {
		iter.next = append(iter.next, next)
	}
	if done != nil {
		iter.done = append(iter.done, done)
	}
}

func (iter *TimeIterator) Next() bool {
	if iter.t >= iter.end {
		return false
	}
	iter.t += iter.step
	for _, next := range iter.next {
		next()
	}
	return true
}

func (iter *TimeIterator) Done() {
	for _, done := range iter.done {
		done()
	}
}

func (iter *TimeIterator) Unix() int64 { return iter.t }

func (iter *TimeIterator) Start() time.Time    { return time.Unix(iter.start, 0) }
func (iter *TimeIterator) End() time.Time      { return time.Unix(iter.end, 0) }
func (iter *TimeIterator) Step() time.Duration { return time.Second * time.Duration(iter.step) }
