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
	MessageStatusRead        = MessageStatus("read")
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

	MessagingServiceSID string `json:"messaging_service_sid"`
}

func (msg *Message) sentMessage() *notification.SentMessage {
	stat := msg.messageStatus()

	return &notification.SentMessage{
		ExternalID:   msg.SID,
		State:        stat.State,
		StateDetails: stat.Details,
		SrcValue:     msg.From,
	}
}

func (msg *Message) messageStatus() *notification.Status {
	if msg == nil {
		return nil
	}

	var status notification.Status
	if msg.ErrorMessage != nil && msg.ErrorCode != nil {
		status.Details = fmt.Sprintf("%s: [%d] %s", msg.Status, *msg.ErrorCode, *msg.ErrorMessage)
	} else {
		status.Details = string(msg.Status)
	}
	switch msg.Status {
	case MessageStatusFailed:
		if msg.ErrorCode != nil &&
			(*msg.ErrorCode == 30008 || *msg.ErrorCode == 30001) {

			status.State = notification.StateFailedTemp
		}
		status.State = notification.StateFailedPerm
	case MessageStatusDelivered:
		status.State = notification.StateDelivered
	case MessageStatusRead:
		status.State = notification.StateRead
	case MessageStatusSent, MessageStatusUndelivered:
		status.State = notification.StateSent
	default:
		status.State = notification.StateSending
	}

	status.SrcValue = msg.From
	return &status
}
