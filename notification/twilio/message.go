package twilio

import (
	"fmt"

	"github.com/target/goalert/notification"
)

// MessageStatus indicates the state of a message.
//
// https://www.twilio.com/docs/api/messaging/message#message-status-values
type MessageStatus string

// Defined status values for messages.
const (
	MessageStatusUnknown     = MessageStatus("")
	MessageStatusAccepted    = MessageStatus("accepted")
	MessageStatusQueued      = MessageStatus("queued")
	MessageStatusSending     = MessageStatus("sending")
	MessageStatusSent        = MessageStatus("sent")
	MessageStatusReceiving   = MessageStatus("receiving")
	MessageStatusReceived    = MessageStatus("received")
	MessageStatusDelivered   = MessageStatus("delivered")
	MessageStatusUndelivered = MessageStatus("undelivered")
	MessageStatusFailed      = MessageStatus("failed")
)

// Scan implements the sql.Scanner interface.
func (s *MessageStatus) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = MessageStatus(t)
	case string:
		*s = MessageStatus(t)
	case nil:
		*s = MessageStatusUnknown
	default:
		return fmt.Errorf("could not process unknown type for Status(%T)", t)
	}
	return nil
}

// A MessageErrorCode is a defined error code for Twilio messages.
//
// https://www.twilio.com/docs/api/messaging/message#delivery-related-errors
type MessageErrorCode int

// Defined error codes for messages.
const (
	MessageErrorCodeQueueOverflow       = MessageErrorCode(30001)
	MessageErrorCodeAccountSuspended    = MessageErrorCode(30002)
	MessageErrorCodeHandsetUnreachable  = MessageErrorCode(30003)
	MessageErrorCodeMessageBlocked      = MessageErrorCode(30004)
	MessageErrorCodeHandsetUnknown      = MessageErrorCode(30005)
	MessageErrorCodeLandlineUnreachable = MessageErrorCode(30006)
	MessageErrorCodeCarrierViolation    = MessageErrorCode(30007)
	MessageErrorCodeUnknown             = MessageErrorCode(30008)
	MessageErrorCodeMissingSegment      = MessageErrorCode(30009)
	MessageErrorCodeExceedsMaxPrice     = MessageErrorCode(30010)
)

// Message represents a Twilio message.
type Message struct {
	SID          string
	To           string
	From         string
	Status       MessageStatus
	ErrorCode    *MessageErrorCode
	ErrorMessage *string
}

func (msg *Message) messageStatus(id string) *notification.MessageStatus {
	if msg == nil {
		return nil
	}

	status := &notification.MessageStatus{
		ID:                id,
		ProviderMessageID: msg.SID,
	}
	if msg.ErrorMessage != nil && msg.ErrorCode != nil {
		status.Details = fmt.Sprintf("%s: [%d] %s", msg.Status, *msg.ErrorCode, *msg.ErrorMessage)
	} else {
		status.Details = string(msg.Status)
	}
	switch msg.Status {
	case MessageStatusFailed:
		if msg.ErrorCode != nil &&
			(*msg.ErrorCode == 30008 || *msg.ErrorCode == 30001) {

			status.State = notification.MessageStateFailedTemp
		}
		status.State = notification.MessageStateFailedPerm
	case MessageStatusDelivered:
		status.State = notification.MessageStateDelivered
	case MessageStatusSent, MessageStatusUndelivered:
		status.State = notification.MessageStateSent
	case MessageStatusAccepted, MessageStatusQueued:
		status.State = notification.MessageStatePending
	default:
		status.State = notification.MessageStateSending
	}
	return status
}
