package twilio

import (
	"sync"
)

const (
	maxPassiveReplyCount = 5
)

type replyLimiter struct {
	mx sync.Mutex

	state map[string]int
}

func newReplyLimiter() *replyLimiter {
	return &replyLimiter{
		state: make(map[string]int),
	}
}

// RecordPassiveReply will increment the number of passive replies to a number.
func (r *replyLimiter) RecordPassiveReply(toNumber string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.state[toNumber]++
}

// ShouldDrop will return true if the message should be dropped.
func (r *replyLimiter) ShouldDrop(toNumber string) bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.state[toNumber] >= maxPassiveReplyCount
}

// Reset will reset the counter for the given number.
func (r *replyLimiter) Reset(toNumber string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.state, toNumber)
}
