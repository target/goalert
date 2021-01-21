package notification

import (
	"database/sql/driver"
	"fmt"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type MessageType

// A Message contains information that can be provided
// to a user for notification.
type Message interface {
	ID() string
	Type() MessageType
	Destination() Dest
}

// MessageType indicates the type of notification message.
type MessageType int

// Allowed types
const (
	MessageTypeUnknown MessageType = iota
	MessageTypeAlert
	MessageTypeAlertStatus
	MessageTypeTest
	MessageTypeVerification
	MessageTypeAlertBundle
	MessageTypeAlertStatusBundle
)

func (s MessageType) Value() (driver.Value, error) {
	switch s {
	case MessageTypeAlert:
		return "alert_notification", nil
	case MessageTypeAlertStatus:
		return "alert_status_update", nil
	case MessageTypeTest:
		return "test_notification", nil
	case MessageTypeVerification:
		return "verification_message", nil
	case MessageTypeAlertBundle:
		return "alert_notification_bundle", nil
	case MessageTypeAlertStatusBundle:
		return "alert_status_update_bundle", nil
	}
	return nil, fmt.Errorf("could not process unknown type for MessageType %s", s)
}

func (s *MessageType) Scan(value interface{}) error {
	str := value.(string)

	switch str {
	case "alert_notification":
		*s = MessageTypeAlert
	case "alert_status_update":
		*s = MessageTypeAlertStatus
	case "test_notification":
		*s = MessageTypeTest
	case "verification_message":
		*s = MessageTypeVerification
	case "alert_notification_bundle":
		*s = MessageTypeAlertBundle
	case "alert_status_update_bundle":
		*s = MessageTypeAlertStatusBundle
	default:
		return fmt.Errorf("could not process unknown type for MessageType %str", str)
	}
	return nil
}
