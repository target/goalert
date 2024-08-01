package nfymsg

import "github.com/target/goalert/gadb"

// MessageType indicates the type of notification message.
type MessageType = gadb.EnumOutgoingMessagesType

// Allowed types
const (
	MessageTypeUnknown      MessageType = ""
	MessageTypeAlert                    = gadb.EnumOutgoingMessagesTypeAlertNotification
	MessageTypeAlertStatus              = gadb.EnumOutgoingMessagesTypeAlertStatusUpdate
	MessageTypeTest                     = gadb.EnumOutgoingMessagesTypeTestNotification
	MessageTypeVerification             = gadb.EnumOutgoingMessagesTypeVerificationMessage
	MessageTypeAlertBundle              = gadb.EnumOutgoingMessagesTypeAlertNotificationBundle

	// MessageTypeAlertStatusBundle is used for bundled status messages.
	//
	// Deprecated: Alert status messages are no longer bundled, status bundle
	// messages are now dropped.
	MessageTypeAlertStatusBundle   = gadb.EnumOutgoingMessagesTypeAlertStatusUpdateBundle
	MessageTypeScheduleOnCallUsers = gadb.EnumOutgoingMessagesTypeScheduleOnCallNotification

	MessageTypeSignalMessage = gadb.EnumOutgoingMessagesTypeSignalMessage
)
