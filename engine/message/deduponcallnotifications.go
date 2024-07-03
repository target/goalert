package message

import (
	"sort"

	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfy"
)

// dedupOnCallNotifications will remove old on-call notifications if a newer one exists for the same schedule & destination.
func dedupOnCallNotifications(messages []Message) ([]Message, []string) {
	toProcess, result := splitPendingByType(messages, notification.MessageTypeScheduleOnCallUsers)
	sort.Slice(toProcess, func(i, j int) bool { return toProcess[i].CreatedAt.After(toProcess[j].CreatedAt) })

	type msgKey struct {
		scheduleID string
		dest       nfy.DestHash
	}

	m := make(map[msgKey]struct{})
	var toDelete []string
	for _, msg := range toProcess {
		key := msgKey{scheduleID: msg.ScheduleID, dest: msg.DestHash()}
		if _, ok := m[key]; ok {
			toDelete = append(toDelete, msg.ID)
			continue
		}

		m[key] = struct{}{}
		result = append(result, msg)
	}

	return result, toDelete
}
