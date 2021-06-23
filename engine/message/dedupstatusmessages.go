package message

import (
	"sort"

	"github.com/target/goalert/notification"
)

// dedupStatusMessages will remove old status updates if a newer one exists for the same alert/destination.
func dedupStatusMessages(messages []Message) ([]Message, []string) {
	sort.Slice(messages, func(i, j int) bool { return messages[i].AlertLogID > messages[j].AlertLogID })

	type msgKey struct {
		alertID int
		dest    notification.Dest
	}

	m := make(map[msgKey]struct{})
	var toDelete []string
	filter := messages[:0]
	for _, msg := range messages {
		//lint:ignore SA1019 TODO delete all occurrences of AlertStatusBundle type and definition
		if msg.Type == notification.MessageTypeAlertStatusBundle {
			// drop old status bundles
			toDelete = append(toDelete, msg.ID)
			continue
		}
		if msg.Type != notification.MessageTypeAlertStatus {
			filter = append(filter, msg)
			continue
		}
		key := msgKey{alertID: msg.AlertID, dest: msg.Dest}
		if _, ok := m[key]; ok {
			toDelete = append(toDelete, msg.ID)
			continue
		}

		m[key] = struct{}{}
		filter = append(filter, msg)
	}

	return filter, toDelete
}
