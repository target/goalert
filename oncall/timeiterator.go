package oncall

import "time"

type TimeIterator struct {
	t, start, end, step int64

	nextStep int64

	next []func(int64) int64
	done []func()
}

type Iterator interface {
	// Process takes the current unix timestamp as a parameter and returns the next
	// actionable step timestamp.
	Process(int64) int64
	Done()
}

func NewTimeIterator(start, end time.Time, step time.Duration) *TimeIterator {
	step = step.Truncate(time.Second)
	stepUnix := step.Nanoseconds() / int64(time.Second)
	start = start.Truncate(step)
	end = end.Truncate(step)

	return &TimeIterator{
		step:     stepUnix,
		start:    start.Unix(),
		end:      end.Unix(),
		nextStep: start.Unix(),
	}
}

func (iter *TimeIterator) Register(next func(int64) int64, done func()) {
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
	iter.t = iter.nextStep
	iter.nextStep = 0

	var nextStep int64
	for _, next := range iter.next {
		nextStep = next(iter.t)
		if nextStep == -1 {
			nextStep = iter.end
		}
		if iter.nextStep == 0 {
			iter.nextStep = nextStep
		} else if nextStep > 0 && nextStep < iter.nextStep {
			iter.nextStep = nextStep
		}
	}
	if iter.nextStep == 0 {
		iter.nextStep = iter.t + iter.step
	} else if iter.nextStep > iter.end {
		iter.nextStep = iter.end
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
