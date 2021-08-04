package twilio

import (
	"sync"
	"time"
)

type numberState struct {
	lastMessage   string
	lastMessageAt time.Time
	errCount      int
}

const (
	maxErrCount = 5
	dupMsgDur   = 5 * time.Minute
)

type replyLimiter struct {
	mx sync.Mutex

	state map[string]numberState
}

func newReplyLimiter() *replyLimiter {
	return &replyLimiter{}
}

// RecordError will record an error for the given number.
func (r *replyLimiter) RecordError(toNumber string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := r.state[toNumber]
	s.errCount++

	r.state[toNumber] = s
}

// RecordAndCheck will return true if the message should be sent.
// It will also record that the message was sent if true.
func (r *replyLimiter) RecordAndCheck(toNumber, message string) bool {
	r.mx.Lock()
	defer r.mx.Unlock()

	s := r.state[toNumber]
	if s.errCount >= maxErrCount {
		return false
	}

	if time.Since(s.lastMessageAt) < dupMsgDur && s.lastMessage == message {
		return false
	}
	s.lastMessage = message
	s.lastMessageAt = time.Now()
	s.errCount = 0
	r.state[toNumber] = s

	return true
}

// Reset will reset the state of the reply limiter for the given number.
func (r *replyLimiter) Reset(toNumber string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.state, toNumber)
}
