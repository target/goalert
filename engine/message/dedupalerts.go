package message

import (
	"sort"

	"github.com/target/goalert/notification"
)

// dedupAlerts will "bundle" identical alert notifications and point them at the oldest pending notification.
func dedupAlerts(msgs []Message, bundleFunc func(parentID string, duplicateIDs []string) error) ([]Message, error) {
	if len(msgs) == 0 {
		return msgs, nil
	}

	sort.Slice(msgs, func(i, j int) bool {
		// if AlertID is the same, then sort by CreatedAt
		if msgs[i].AlertID == msgs[j].AlertID {
			return msgs[i].CreatedAt.Before(msgs[j].CreatedAt)
		}
		return msgs[i].AlertID < msgs[j].AlertID
	})

	type msgKey struct {
		notification.Dest
		AlertID int
	}
	alerts := make(map[msgKey]string, len(msgs))
	duplicates := make(map[string][]string)

	filtered := msgs[:0]
	for _, msg := range msgs {
		if msg.Type != notification.MessageTypeAlert {
			// skip non-alert messages
			filtered = append(filtered, msg)
			continue
		}
		if !msg.SentAt.IsZero() {
			// skip sent messages
			filtered = append(filtered, msg)
			continue
		}

		// check if we have seen this alert before
		key := msgKey{msg.Dest, msg.AlertID}

		if parentID, ok := alerts[key]; ok {
			duplicates[parentID] = append(duplicates[parentID], msg.ID)
			continue
		}

		alerts[key] = msg.ID
		filtered = append(filtered, msg)
	}

	for parentID, duplicateIDs := range duplicates {
		err := bundleFunc(parentID, duplicateIDs)
		if err != nil {
			return nil, err
		}
	}

	return filtered, nil
}
