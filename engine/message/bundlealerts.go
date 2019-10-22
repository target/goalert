package message

import (
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/notification"
)

// bundleAlertMessages will bundle status updates for the same Dest value. It will add any new messages to an existing bundle.
// A single contact-method will only ever have a single alert notification per-service in the result.
//
// It also handles updating the outgoing_messages table by marking bundled messages with the `bundled`
// status and creating a new bundled message placeholder.
func bundleAlertMessages(messages []Message, bundleFunc func(Message, []string) error) ([]Message, error) {
	type bundleID struct {
		notification.Dest
		ServiceID string
	}
	type bundle struct {
		Message
		IDs []string
	}
	byID := make(map[bundleID]*bundle)
	filtered := messages[:0]
	for _, msg := range messages {
		if (msg.Type != TypeAlertNotification && msg.Type != TypeAlertNotificationBundle) || !msg.SentAt.IsZero() {
			filtered = append(filtered, msg)
			continue
		}

		id := bundleID{Dest: msg.Dest, ServiceID: msg.ServiceID}
		old, ok := byID[id]
		if !ok {
			cpy := bundle{Message: msg, IDs: []string{msg.ID}}
			byID[id] = &cpy
			continue
		}

		// duplicate alert
		old.IDs = append(old.IDs, msg.ID)

		if msg.CreatedAt.Before(old.CreatedAt) {
			old.CreatedAt = msg.CreatedAt
		}
	}

	for _, msg := range byID {
		if len(msg.IDs) == 1 {
			filtered = append(filtered, msg.Message)
			continue
		}

		msg.Type = TypeAlertNotificationBundle
		msg.AlertID = 0
		msg.ID = uuid.NewV4().String()
		err := bundleFunc(msg.Message, msg.IDs)
		if err != nil {
			return nil, err
		}
		filtered = append(filtered, msg.Message)
	}

	return filtered, nil
}
