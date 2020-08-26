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
	MessageTypeAlert MessageType = iota
	MessageTypeAlertStatus
	MessageTypeTest
	MessageTypeVerification
	MessageTypeAlertBundle
	MessageTypeAlertStatusBundle
)

func (s MessageType) Value() (driver.Value, error) {
	switch s {
	case 0:
		return "alert_notification", nil
	case 1:
		return "alert_status_update", nil
	case 2:
		return "test_notification", nil
	case 3:
		return "verification_message", nil
	case 4:
		return "alert_notification_bundle", nil
	case 5:
		return "alert_status_update_bundle", nil
	}
	return nil, fmt.Errorf("could not process unknown type for %s", s)
}

func (s *MessageType) Scan(value interface{}) error {
	switch t := value.(type) {
	case int:
		*s = MessageType(t)
	default:
		return fmt.Errorf("could not process unknown type %t", t)
	}
	return nil
}