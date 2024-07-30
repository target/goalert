package message

import (
	"sort"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
)

// dedupStatusMessages will remove old status updates if a newer one exists for the same alert/destination.
func dedupStatusMessages(messages []Message) ([]Message, []string) {
	toProcess, result := splitPendingByType(messages, notification.MessageTypeAlertStatus)
	sort.Slice(toProcess, func(i, j int) bool { return toProcess[i].AlertLogID > toProcess[j].AlertLogID })

	type msgKey struct {
		alertID int
		dest    gadb.DestHash
	}

	m := make(map[msgKey]struct{})
	var toDelete []string
	for _, msg := range toProcess {
		key := msgKey{alertID: msg.AlertID, dest: msg.Dest.DestHash()}
		if _, ok := m[key]; ok {
			toDelete = append(toDelete, msg.ID)
			continue
		}

		m[key] = struct{}{}
		result = append(result, msg)
	}

	return result, toDelete
}
