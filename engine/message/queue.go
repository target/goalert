package message

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/target/goalert/notification"
)

var typePriority = map[notification.MessageType]int{
	notification.MessageTypeVerification: 1,
	notification.MessageTypeTest:         2,

	notification.MessageTypeScheduleOnCallUsers: 3,

	// First alert will jump the list with priority 0, so this only
	// represents additional alerts to the service after the first.
	notification.MessageTypeAlert:       4,
	notification.MessageTypeAlertBundle: 4,

	notification.MessageTypeAlertStatus: 5,

	notification.MessageTypeSignalMessage: 99, // lowest priority
}

type queue struct {
	sent    []Message
	pending map[string][]Message
	now     time.Time

	firstAlert  map[destID]struct{}
	serviceSent map[string]time.Time
	userSent    map[string]time.Time
	destSent    map[notification.DestID]time.Time

	cmThrottle     *Throttle
	globalThrottle *Throttle

	mx sync.Mutex
}

type destID struct {
	ID       string
	DestType string
}

func newQueue(msgs []Message, now time.Time) *queue {
	q := &queue{
		sent:    make([]Message, 0, len(msgs)),
		pending: make(map[string][]Message),
		now:     now,

		firstAlert:  make(map[destID]struct{}),
		serviceSent: make(map[string]time.Time),
		userSent:    make(map[string]time.Time),
		destSent:    make(map[notification.DestID]time.Time),

		cmThrottle:     NewThrottle(PerCMThrottle, now, false),
		globalThrottle: NewThrottle(GlobalCMThrottle, now, true),
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

	q.cmThrottle.Record(m)
	q.globalThrottle.Record(m)
	q.firstAlert[destID{ID: m.ServiceID, DestType: m.Dest.Type}] = struct{}{}
	if t := q.serviceSent[m.ServiceID]; m.SentAt.After(t) {
		q.serviceSent[m.ServiceID] = m.SentAt
	}
	if t := q.userSent[m.UserID]; m.SentAt.After(t) {
		q.userSent[m.UserID] = m.SentAt
	}
	if t := q.destSent[m.DestID]; m.SentAt.After(t) {
		q.destSent[m.DestID] = m.SentAt
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
		// neither has received a message
		return false, false
	}

	return sentA.Before(sentB), true
}

// filterPending will delete messages from pending that are not eligible to be sent.
func (q *queue) filterPending(destType string) {
	pending := q.pending[destType]
	if len(pending) == 0 {
		return
	}

	filtered := pending[:0]
	for _, p := range pending {
		if q.globalThrottle.InCooldown(p) {
			continue
		}
		if q.cmThrottle.InCooldown(p) {
			continue
		}
		filtered = append(filtered, p)
	}

	q.pending[destType] = filtered
}

// sortPending will re-sort the list of pending messages.
func (q *queue) sortPending(destType string) {
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
		_, firstAlertI := q.firstAlert[destID{ID: pi.ServiceID, DestType: pi.Dest.Type}]
		if (pi.Type == notification.MessageTypeAlert || pi.Type == notification.MessageTypeAlertBundle) && !firstAlertI {
			piTypePriority = 0
		}

		pjTypePriority := typePriority[pj.Type]
		_, firstAlertJ := q.firstAlert[destID{ID: pj.ServiceID, DestType: pj.Dest.Type}]
		if (pj.Type == notification.MessageTypeAlert || pj.Type == notification.MessageTypeAlertBundle) && !firstAlertJ {
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
func (q *queue) NextByType(destType string) *Message {
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
func (q *queue) SentByType(destType string, dur time.Duration) int {
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
func (q *queue) Types() []string {
	q.mx.Lock()
	defer q.mx.Unlock()

	result := make([]string, 0, len(q.pending))
	for typ, msgs := range q.pending {
		if len(msgs) == 0 {
			continue
		}

		result = append(result, typ)
	}

	return result
}
