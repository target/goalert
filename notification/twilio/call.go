package twilio

import (
	"fmt"
	"time"

	"github.com/target/goalert/notification"
)

// CallStatus indicates the state of a voice call.
//
// https://www.twilio.com/docs/api/twiml/twilio_request#request-parameters-call-status
type CallStatus string

// Defined status values for voice calls.
const (
	CallStatusUnknown    = CallStatus("")
	CallStatusInitiated  = CallStatus("initiated")
	CallStatusQueued     = CallStatus("queued")
	CallStatusRinging    = CallStatus("ringing")
	CallStatusInProgress = CallStatus("in-progress")
	CallStatusCompleted  = CallStatus("completed")
	CallStatusBusy       = CallStatus("busy")
	CallStatusFailed     = CallStatus("failed")
	CallStatusNoAnswer   = CallStatus("no-answer")
	CallStatusCanceled   = CallStatus("canceled")
)

// Scan implements the sql.Scanner interface.
func (s *CallStatus) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = CallStatus(t)
	case string:
		*s = CallStatus(t)
	case nil:
		*s = CallStatusUnknown
	default:
		return fmt.Errorf("could not process unknown type for Status(%T)", t)
	}
	return nil
}

// CallErrorCode is an error code encountered when making a call.
type CallErrorCode int

// Call represents a Twilio voice call.
type Call struct {
	SID            string
	To             string
	From           string
	Status         CallStatus
	SequenceNumber *int
	Direction      string
	CallDuration   time.Duration
	ErrorMessage   *string
	ErrorCode      *CallErrorCode

	MessagingServiceSID string `json:"messaging_service_sid"`
}

func (call *Call) sentMessage() *notification.SentMessage {
	stat := call.messageStatus()

	return &notification.SentMessage{
		ExternalID:   call.SID,
		State:        stat.State,
		StateDetails: stat.Details,
		SrcValue:     call.From,
	}
}
func (call *Call) messageStatus() *notification.Status {
	if call == nil {
		return nil
	}

	var status notification.Status
	if call.ErrorMessage != nil && call.ErrorCode != nil {
		status.Details = fmt.Sprintf("%s: [%d] %s", call.Status, *call.ErrorCode, *call.ErrorMessage)
	} else {
		status.Details = string(call.Status)
	}
	if call.SequenceNumber != nil {
		status.Sequence = *call.SequenceNumber
	}

	switch call.Status {
	case CallStatusCompleted:
		status.State = notification.StateDelivered
	case CallStatusInitiated, CallStatusQueued:
		status.State = notification.StateSending
	case CallStatusBusy:
		status.State = notification.StateFailedTemp
	case CallStatusFailed, CallStatusCanceled, CallStatusNoAnswer:
		status.State = notification.StateFailedPerm
	default:
		status.State = notification.StateSent
	}
	return &status
}
