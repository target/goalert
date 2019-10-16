package message

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/target/goalert/notification"
)

// cmCooldown is the amount of time (minimum) between messages to a particular contact method.
const cmCooldown = time.Minute

var typePriority = map[Type]int{
	TypeVerificationMessage: 1,
	TypeTestNotification:    2,
	TypeAlertNotification:   3,
	TypeAlertStatusUpdate:   4,
}

type queue struct {
	sent    []Message
	pending map[notification.DestType][]Message
	now     time.Time

	serviceSent map[string]time.Time
	userSent    map[string]time.Time
	destSent    map[notification.Dest]time.Time

	mx sync.Mutex
}

func newQueue(msgs []Message, now time.Time) *queue {
	q := &queue{
		sent:    make([]Message, 0, len(msgs)),
		pending: make(map[notification.DestType][]Message),
		now:     now,

		serviceSent: make(map[string]time.Time),
		userSent:    make(map[string]time.Time),
		destSent:    make(map[notification.Dest]time.Time),
	}

	for _, m := range msgs {
		if m.SentAt.IsZero() {
			q.pending[m.Dest.Type] = append(q.pending[m.Dest.Type], m)
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
		q.userSent[m.UserID] = m.SentAt
	}
	if t := q.destSent[m.Dest]; m.SentAt.After(t) {
		q.destSent[m.Dest] = m.SentAt
	}

	q.sent = append(q.sent, m)
}

func (q *queue) userPriority(userA, userB string) (isLess, ok bool) {
	sentA := q.userSent[userA]
	sentB := q.userSent[userB]

	if sentA.Equal(sentB) {
		return false, false
	}

	return sentA.Before(sentB), true
}

func (q *queue) servicePriority(serviceA, serviceB string) (isLess, ok bool) {
	sentA := q.serviceSent[serviceA]
	sentB := q.serviceSent[serviceB]

	if sentA.Equal(sentB) {
		// neither has recieved a message
		return false, false
	}

	return sentA.Before(sentB), true
}

// filterPending will delete messages from pending that are not eligible to be sent.
func (q *queue) filterPending(destType notification.DestType) {
	pending := q.pending[destType]
	if len(pending) == 0 {
		return
	}

	cutoffTime := q.now.Add(-cmCooldown)
	filtered := pending[:0]
	for _, p := range pending {
		if q.destSent[p.Dest].After(cutoffTime) {
			continue
		}
		filtered = append(filtered, p)
	}

	q.pending[destType] = filtered
}

// sortPending will re-sort the list of pending messages.
func (q *queue) sortPending(destType notification.DestType) {
	pending := q.pending[destType]
	if len(pending) == 0 {
		return
	}

	rand.Shuffle(len(pending), func(i, j int) { pending[i], pending[j] = pending[j], pending[i] })
	sort.SliceStable(pending, func(i, j int) bool {
		pi, pj := pending[i], pending[j]
		if pi.CreatedAt.Equal(pj.CreatedAt) {
			// keep existing order
			return i < j
		}
		// sort by creation time (if timestamps are not equal)
		return pi.CreatedAt.Before(pj.CreatedAt)
	})
	sort.SliceStable(pending, func(i, j int) bool {
		pi, pj := pending[i], pending[j]

		// First Alert to a service takes highest priority
		piTypePriority := typePriority[pi.Type]
		if pi.Type == TypeAlertNotification && q.serviceSent[pi.ServiceID].IsZero() {
			piTypePriority = 0
		}

		pjTypePriority := typePriority[pj.Type]
		if pj.Type == TypeAlertNotification && q.serviceSent[pj.ServiceID].IsZero() {
			pjTypePriority = 0
		}

		if piTypePriority != pjTypePriority {
			return piTypePriority < pjTypePriority
		}

		if isLess, ok := q.userPriority(pi.UserID, pj.UserID); ok {
			return isLess
		}

		if isLess, ok := q.servicePriority(pi.ServiceID, pj.ServiceID); ok {
			return isLess
		}

		// two different users, two different services, none have gotten any notification
		// return false to keep random ordering
		return i < j
	})

	q.pending[destType] = pending
}

// NextByType returns the next message to be sent
// for the given type.
//
// It returns nil if there are no more messages.
func (q *queue) NextByType(destType notification.DestType) *Message {
	q.mx.Lock()
	defer q.mx.Unlock()

	q.filterPending(destType)
	q.sortPending(destType)
	pending := q.pending[destType]
	if len(pending) == 0 {
		return nil
	}

	next := pending[0]
	q.pending[destType] = pending[1:]
	q.addSent(next)

	return &next
}

// SentByType returns the number of messages sent for the given type
// over the past Duration.
func (q *queue) SentByType(destType notification.DestType, dur time.Duration) int {
	q.mx.Lock()
	defer q.mx.Unlock()

	cutoff := q.now.Add(-dur)

	var count int
	for _, msg := range q.sent {
		if msg.SentAt.After(cutoff) && msg.Dest.Type == destType {
			count++
		}
	}
	return count
}

// Types returns a slice of all DestTypes currently waiting to be sent.
func (q *queue) Types() []notification.DestType {
	q.mx.Lock()
	defer q.mx.Unlock()

	result := make([]notification.DestType, 0, len(q.pending))
	for typ, msgs := range q.pending {
		if len(msgs) == 0 {
			continue
		}

		result = append(result, typ)
	}

	return result
}
