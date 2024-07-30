package message

import (
	"sort"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
)

// dedupAlerts will "bundle" identical alert notifications and point them at the oldest pending notification.
func dedupAlerts(msgs []Message, bundleFunc func(parentID string, duplicateIDs []string) error) ([]Message, error) {
	if len(msgs) == 0 {
		return msgs, nil
	}

	toProcess, result := splitPendingByType(msgs, notification.MessageTypeAlert)

	// sort by "created" time
	sort.Slice(toProcess, func(i, j int) bool { return toProcess[i].CreatedAt.Before(toProcess[j].CreatedAt) })

	type msgKey struct {
		gadb.DestHash
		AlertID int
	}
	alerts := make(map[msgKey]string, len(msgs))
	duplicates := make(map[string][]string)

	for _, msg := range toProcess {
		// check if we have seen this alert before
		key := msgKey{msg.Dest.DestHash(), msg.AlertID}

		if parentID, ok := alerts[key]; ok {
			duplicates[parentID] = append(duplicates[parentID], msg.ID)
			continue
		}

		alerts[key] = msg.ID
		result = append(result, msg)
	}

	for parentID, duplicateIDs := range duplicates {
		err := bundleFunc(parentID, duplicateIDs)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
