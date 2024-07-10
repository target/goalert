package notification

import (
	"database/sql/driver"
	"fmt"

	"github.com/target/goalert/gadb"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type MessageType

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

	// MessageTypeAlertStatusBundle is used for bundled status messages.
	//
	// Deprecated: Alert status messages are no longer bundled, status bundle
	// messages are now dropped.
	MessageTypeAlertStatusBundle
	MessageTypeScheduleOnCallUsers

	MessageTypeSignalMessage
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
	case MessageTypeScheduleOnCallUsers:
		return "schedule_on_call_notification", nil
	case MessageTypeSignalMessage:
		return "signal_message", nil
	}
	return nil, fmt.Errorf("could not process unknown type for MessageType %s", s)
}

func (s *MessageType) FromDB(value gadb.EnumOutgoingMessagesType) error { return s.Scan(string(value)) }

func (s MessageType) ToDB() (gadb.EnumOutgoingMessagesType, error) {
	val, err := s.Value()
	if err != nil {
		return "", err
	}
	return gadb.EnumOutgoingMessagesType(val.(string)), nil
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
	case "schedule_on_call_notification":
		*s = MessageTypeScheduleOnCallUsers
	case "signal_message":
		*s = MessageTypeSignalMessage
	default:
		return fmt.Errorf("could not process unknown type for MessageType %str", str)
	}
	return nil
}
