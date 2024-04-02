package message

import (
	"sort"

	"github.com/target/goalert/notification"
)

// dedupStatusMessages will remove old status updates if a newer one exists for the same alert/destination.
func dedupStatusMessages(messages []Message) ([]Message, []string) {
	toProcess, result := splitPendingByType(messages, notification.MessageTypeAlertStatus)
	sort.Slice(toProcess, func(i, j int) bool {
		iMsg, jMsg := toProcess[i], toProcess[j]

		if iMsg.AlertLogID == 0 || jMsg.AlertLogID == 0 {
			// sort by created at if no alert log ID
			return toProcess[i].CreatedAt.After(toProcess[j].CreatedAt)
		}

		return toProcess[i].AlertLogID > toProcess[j].AlertLogID
	})

	type msgKey struct {
		alertID int
		dest    notification.Dest
	}

	m := make(map[msgKey]struct{})
	var toDelete []string
	for _, msg := range toProcess {
		key := msgKey{alertID: msg.AlertID, dest: msg.Dest}
		if _, ok := m[key]; ok {
			toDelete = append(toDelete, msg.ID)
			continue
		}

		m[key] = struct{}{}
		result = append(result, msg)
	}

	return result, toDelete
}
