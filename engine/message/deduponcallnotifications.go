package message

import (
	"sort"

	"github.com/target/goalert/notification"
)

// dedupOnCallNotifications will remove old on-call notifications if a newer one exists for the same schedule & destination.
func dedupOnCallNotifications(messages []Message) ([]Message, []string) {
	sort.Slice(messages, func(i, j int) bool { return messages[i].CreatedAt.After(messages[j].CreatedAt) })

	type msgKey struct {
		scheduleID string
		dest       notification.Dest
	}

	m := make(map[msgKey]struct{})
	var toDelete []string
	filter := messages[:0]
	for _, msg := range messages {
		if msg.Type != notification.MessageTypeScheduleOnCallUsers {
			filter = append(filter, msg)
			continue
		}
		key := msgKey{scheduleID: msg.ScheduleID, dest: msg.Dest}
		if _, ok := m[key]; ok {
			toDelete = append(toDelete, msg.ID)
			continue
		}

		m[key] = struct{}{}
		filter = append(filter, msg)
	}

	return filter, toDelete
}
