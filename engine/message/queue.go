package message

import (
	"math/rand"
	"sort"
	"time"

	"github.com/target/goalert/notification"
)

// cmCooldown is the amount of time (minimum) between messages to a particular contact method.
const cmCooldown = time.Minute

var typePriority = map[Type]int{
	TypeAlertNotification:   0, // First Alert
	TypeVerificationMessage: 1,
	TypeTestNotification:    2,

	// Non-first Alert for Same user
	TypeAlertStatusUpdate: 3,
}

type queue struct {
	sent    []Message
	pending []Message
	now     time.Time

	serviceSent map[string]time.Time
	userSent    map[string]time.Time
	destSent    map[notification.Dest]time.Time
}

func newQueue(msgs []Message, now time.Time) *queue {
	q := &queue{
		sent:    make([]Message, 0, len(msgs)),
		pending: make([]Message, 0, len(msgs)),
		now:     now,

		serviceSent: make(map[string]time.Time),
		userSent:    make(map[string]time.Time),
		destSent:    make(map[notification.Dest]time.Time),
	}

	for _, m := range msgs {
		if m.SentAt.IsZero() {
			q.pending = append(q.pending, m)
			continue
		}
		q.addSent(m)
	}

	return q
}
func (q *queue) addSent(m Message) {
	if m.SentAt.IsZero() {
		m.SentAt = q.now
	}

	if t := q.serviceSent[m.ServiceID]; m.SentAt.After(t) {
		q.serviceSent[m.ServiceID] = m.SentAt
	}
	if t := q.userSent[m.UserID]; m.SentAt.After(t) {
		q.serviceSent[m.UserID] = m.SentAt
	}
	if t := q.destSent[m.Dest]; m.SentAt.After(t) {
		q.destSent[m.Dest] = m.SentAt
	}

	q.sent = append(q.sent, m)
}

func (q *queue) userPriority(userA, userB string) (isLess, ok bool) {
	sentA := q.userSent[userA]
	sentB := q.userSent[userB]

	if sentA.IsZero() && sentB.IsZero() {
		// neither has recieved a message
		return false, false
	}

	return sentA.Before(sentB), true
}

func (q *queue) servicePriority(serviceA, serviceB string) (isLess, ok bool) {
	sentA := q.serviceSent[serviceA]
	sentB := q.serviceSent[serviceB]

	if sentA.IsZero() && sentB.IsZero() {
		// neither has recieved a message
		return false, false
	}

	return sentA.Before(sentB), true
}

// filterPending will delete messages from pending that are not eligible to be sent.
func (q *queue) filterPending() {
	cutoffTime := q.now.Add(-cmCooldown)
	filtered := q.pending[:0]
	for _, p := range q.pending {
		if q.destSent[p.Dest].After(cutoffTime) {
			continue
		}
		filtered = append(filtered, p)
	}

	q.pending = filtered
}

// sortPending will re-sort the list of pending messages.
func (q *queue) sortPending() {
	rand.Shuffle(len(q.pending), func(i, j int) { q.pending[i], q.pending[j] = q.pending[j], q.pending[i] })
	sort.SliceStable(q.pending, func(i, j int) bool {
		pi, pj := q.pending[i], q.pending[j]
		// sort by creation time (if timestamps are not equal)
		return pi.CreatedAt.Before(pj.CreatedAt)
	})
	sort.SliceStable(q.pending, func(i, j int) bool {
		pi, pj := q.pending[i], q.pending[j]
		if pi.Type != pj.Type {
			return typePriority[pi.Type] < typePriority[pj.Type]
		}

		if isLess, ok := q.userPriority(pi.UserID, pj.UserID); ok {
			return isLess
		}

		if isLess, ok := q.servicePriority(pi.ServiceID, pj.ServiceID); ok {
			return isLess
		}

		// two different users, two different services, none have gotten any notification
		// return false to keep random ordering
		return false
	})
}

// NextByType returns the next message to be sent
// for the given type.
//
// It returns nil if there are no more messages.
func (q *queue) NextByType(destType notification.DestType) *Message {
	q.filterPending()
	q.sortPending()
	if len(q.pending) == 0 {
		return nil
	}

	next := q.pending[0]
	q.pending = q.pending[1:]
	q.addSent(next)

	return &next
}

// SentByType returns the number of messages sent for the given type
// over the past Duration.
func (q *queue) SentByType(destType notification.DestType, dur time.Duration) int {
	var count int
	for _, msg := range q.sent {
		if msg.Dest.Type == destType {
			count++
		}
	}
	return count
}

// Types returns a slice of all DestTypes currently waiting to be sent.
func (q *queue) Types() []notification.DestType {
	types := make(map[notification.DestType]bool)
	for _, p := range q.pending {
		types[p.Dest.Type] = true
	}

	result := make([]notification.DestType, 0, len(types))
	for typ := range types {
		result = append(result, typ)
	}

	return result
}
