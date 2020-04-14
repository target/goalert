package alertlog

import "fmt"

// A Type represents a log entry type for an alert.
type Type string

// Types of Log Entries
const (
	TypeCreated            Type = "created"
	TypeClosed             Type = "closed"
	TypeNotificationSent   Type = "notification_sent"
	TypeNoNotificationSent Type = "no_notification_sent"
	TypeEscalated          Type = "escalated"
	TypeAcknowledged       Type = "acknowledged"
	TypePolicyUpdated      Type = "policy_updated"
	TypeDuplicateSupressed Type = "duplicate_suppressed"
	TypeEscalationRequest  Type = "escalation_request"

	// not exported, status_changed will be turned into an acknowledged where appropriate
	_TypeStatusChanged Type = "status_changed"

	// not exported, response_received will be turned into an ack or closed
	_TypeResponseReceived Type = "response_received"

	// Mapped to Ack and Close
	_TypeAcknowledgeAll Type = "ack_all"
	_TypeCloseAll       Type = "close_all"
)

// Scan handles reading a Type from the DB enum
func (ty *Type) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*ty = Type(t)
	case string:
		*ty = Type(t)
	default:
		return fmt.Errorf("could not process unknown type %T", t)
	}

	return nil
}
