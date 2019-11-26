package message

// Type represents the purpose of a message in the outgoing messages queue.
type Type string

// defined message types
const (
	TypeAlertNotification       Type = "alert_notification"
	TypeTestNotification        Type = "test_notification"
	TypeVerificationMessage     Type = "verification_message"
	TypeAlertStatusUpdate       Type = "alert_status_update"
	TypeAlertNotificationBundle Type = "alert_notification_bundle"
	TypeAlertStatusUpdateBundle Type = "alert_status_update_bundle"
)
