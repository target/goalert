package oncall

import (
	"fmt"
	"time"
)

// TimeIterator will iterate between start and end at a particular step interval.
type TimeIterator struct {
	t, start, end, step int64

	nextStep int64
	init     bool

	sub []SubIterator
}

// A SubIterator can be added to a TimeIterator via the Register method.
type SubIterator interface {
	// Process will be called with each timestamp that needs processing. Each call will be sequential, but
	// it may not be called for each step.
	//
	// Process should return the value of the next required timestamp if it is known otherwise 0.
	// If the iterator has no more events to process -1 can be returned to signal complete.
	//
	// The value returned by Process must be -1, 0, or greater than t.
	Process(t int64) int64

	// Done will be called when the iterator is no longer needed.
	Done()
}

// Startable is a SubIterator method that allows reporting an alternate start timestamp that may extend beyond the parent TimeIterator.
type Startable interface {
	// StartUnix should return the earliest start time for this SubIterator. If it is unknown 0 should be returned.
	StartUnix() int64
}

// NextFunc can be used as a SubIterator.
type NextFunc func(int64) int64

// Process implements the SubIterator.Process method by calling the NextFunc.
func (fn NextFunc) Process(t int64) int64 { return fn(t) }

// Done is just a stub to implement the SubIterator.Done method.
func (fn NextFunc) Done() {}

// NewTimeIterator will create a new TimeIterator with the given configuration.
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

// Register adds a new sub Iterator.
func (iter *TimeIterator) Register(sub SubIterator) { iter.sub = append(iter.sub, sub) }

// Next will return true until iteration completes.
func (iter *TimeIterator) Next() bool {
	if !iter.init {
		for _, s := range iter.sub {
			st, ok := s.(Startable)
			if !ok {
				continue
			}
			start := st.StartUnix()
			start = start - start%iter.step
			if start != 0 && start < iter.start {
				iter.start = start
			}

		}
		iter.nextStep = iter.start
		iter.init = true
	}
	if iter.t >= iter.end {
		return false
	}
	iter.t = iter.nextStep
	iter.nextStep = 0

	var nextStep int64
	for _, sub := range iter.sub {
		nextStep = sub.Process(iter.t)
		if nextStep > 0 && nextStep <= iter.t {
			panic(fmt.Sprintf("nextStep was not in the future; got %d; want > %d (start=%d, end=%d)\n%#v", nextStep, iter.t, iter.start, iter.end, sub))
		}
		if nextStep == -1 {
			// -1 means nothing left, jump to end
			nextStep = iter.end
		}
		if iter.nextStep == 0 {
			// no hints yet, so use current one
			iter.nextStep = nextStep
		} else if nextStep > 0 && nextStep < iter.nextStep {
			// we have a next timestamp, update it only if it's sooner than the current hint
			iter.nextStep = nextStep
		}
	}

	if iter.nextStep == 0 {
		// no hints, so proceed to the next step
		iter.nextStep = iter.t + iter.step
	} else if iter.nextStep > iter.end {
		// clamp nextStep to end
		iter.nextStep = iter.end
	}
	return true
}

// Close should be called when the iterator is no longer needed.
func (iter *TimeIterator) Close() {
	for _, s := range iter.sub {
		s.Done()
	}
}

// Unix will return the current unix timestamp (seconds).
func (iter *TimeIterator) Unix() int64 { return iter.t }

// Start will return start time of the TimeIterator.
func (iter *TimeIterator) Start() time.Time { return time.Unix(iter.start, 0) }

// End will return the end time of the TimeIterator.
func (iter *TimeIterator) End() time.Time { return time.Unix(iter.end, 0) }

// Step will return the iterators step value.
func (iter *TimeIterator) Step() time.Duration { return time.Second * time.Duration(iter.step) }
