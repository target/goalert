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
	toProcess, result := splitPendingByType(messages, notification.MessageTypeAlert, notification.MessageTypeAlertBundle)
	// sort by type, then CreatedAt
	sort.Slice(toProcess, func(i, j int) bool {
		if toProcess[i].Type != toProcess[j].Type {
			return typeOrder(toProcess[i]) < typeOrder(toProcess[j])
		}
		return toProcess[i].CreatedAt.Before(toProcess[j].CreatedAt)
	})

	type key struct {
		notification.DestID
		ServiceID string
	}

	groups := make(map[key][]Message)
	for _, msg := range toProcess {
		key := key{
			DestID:    msg.DestID,
			ServiceID: msg.ServiceID,
		}
		groups[key] = append(groups[key], msg)
	}

	for _, msgs := range groups {
		// skip single messages
		if len(msgs) == 1 {
			result = append(result, msgs[0])
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
		result = append(result, msgs[0])
	}

	return result, nil
}
