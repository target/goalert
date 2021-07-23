package message

import (
	"sort"

	"github.com/target/goalert/notification"
)

func typeOrder(msg Message) int {
	if msg.Type == notification.MessageTypeAlertBundle {
		return 0
	}

	return 1
}

// bundleAlertMessages will bundle status updates for the same Dest value. It will add any new messages to an existing bundle.
// A single contact-method will only ever have a single alert notification per-service in the result.
//
// It also handles updating the outgoing_messages table by marking bundled messages with the `bundled`
// status and creating a new bundled message placeholder.
func bundleAlertMessages(messages []Message, newBundleFunc func(Message) (string, error), bundleFunc func(string, []string) error) ([]Message, error) {
	type key struct {
		notification.Dest
		ServiceID string
	}

	// sort by type, then CreatedAt
	sort.Slice(messages, func(i, j int) bool {
		if messages[i].Type != messages[j].Type {
			return typeOrder(messages[i]) < typeOrder(messages[j])
		}
		return messages[i].CreatedAt.Before(messages[j].CreatedAt)
	})

	groups := make(map[key][]Message)

	filtered := messages[:0]
	for _, msg := range messages {
		// ignore messages that have been sent
		if !msg.SentAt.IsZero() {
			filtered = append(filtered, msg)
			continue
		}

		// ignore anything that is not an alert or alert bundle
		if msg.Type != notification.MessageTypeAlert && msg.Type != notification.MessageTypeAlertBundle {
			filtered = append(filtered, msg)
			continue
		}

		key := key{
			Dest:      msg.Dest,
			ServiceID: msg.ServiceID,
		}
		groups[key] = append(groups[key], msg)
	}

	for _, msgs := range groups {
		// skip single messages
		if len(msgs) == 1 {
			filtered = append(filtered, msgs[0])
			continue
		}

		ids := make([]string, len(msgs))
		for i, msg := range msgs {
			ids[i] = msg.ID
		}

		bundleID := msgs[0].ID
		var err error
		if msgs[0].Type != notification.MessageTypeAlertBundle {
			bundleID, err = newBundleFunc(msgs[0])
			if err != nil {
				return nil, err
			}
			msgs[0].Type = notification.MessageTypeAlertBundle
			msgs[0].ID = bundleID
			msgs[0].AlertID = 0
		} else {
			ids = ids[1:]
		}

		err = bundleFunc(bundleID, ids)
		if err != nil {
			return nil, err
		}
		filtered = append(filtered, msgs[0])
	}

	return filtered, nil
}
