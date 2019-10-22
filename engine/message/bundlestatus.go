package message

import (
	"sort"

	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/notification"
)

// bundleStatusMessages will bundle status updates for the same Dest value. It will not attempt to join existing status-update bundles.
//
// It also handles updating the outgoing_messages table by marking bundled messages with the `bundled`
// status and creating a new bundled message placeholder.
func bundleStatusMessages(messages []Message, bundleFunc func(Message, []string) error) ([]Message, error) {
	sort.Slice(messages, func(i, j int) bool { return messages[i].AlertLogID > messages[j].AlertLogID })
	type bundle struct {
		Message
		IDs []string
	}
	byDest := make(map[notification.Dest]*bundle)
	filtered := messages[:0]
	for _, msg := range messages {
		if msg.Type != TypeAlertStatusUpdate || !msg.SentAt.IsZero() {
			filtered = append(filtered, msg)
			continue
		}

		old, ok := byDest[msg.Dest]
		if !ok {
			cpy := bundle{Message: msg, IDs: []string{msg.ID}}
			byDest[msg.Dest] = &cpy
			continue
		}

		old.IDs = append(old.IDs, msg.ID)
		if msg.CreatedAt.Before(old.CreatedAt) {
			// use oldest value as CreatedAt
			old.CreatedAt = msg.CreatedAt
		}
	}

	// add Bundled messages to the table
	for _, msg := range byDest {
		if len(msg.IDs) == 1 {
			filtered = append(filtered, msg.Message)
			continue
		}

		msg.Type = TypeAlertStatusUpdateBundle
		msg.ID = uuid.NewV4().String()
		msg.StatusCount = len(msg.IDs)
		err := bundleFunc(msg.Message, msg.IDs)
		if err != nil {
			return nil, err
		}
		filtered = append(filtered, msg.Message)
	}

	return filtered, nil
}
