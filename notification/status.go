package notification

import (
	"fmt"
	"time"
)

// Status describes the current state of an outgoing message.
type Status struct {

	// State is the current state.
	State State

	// Details can contain any additional information about the State (e.g. "ringing", "no-answer" etc..).
	Details string

	// Sequence can be used when the provider sends updates out-of order (e.g. Twilio).
	// The Sequence number defaults to 0, and a status update is ignored unless its
	// Sequence number is >= the current one.
	Sequence int

	// SrcValue can be used to set/update the source value of the message.
	SrcValue string

	age time.Duration
}

// Age returns the amount of time from the message creation until the last status update, if available.
func (s Status) Age() time.Duration { return s.age }

// SendResult represents the result of a sent message.
type SendResult struct {
	// ID is the GoAlert message ID.
	ID string

	// ProviderMessageID is an identifier that represents the provider-specific ID
	// of the message (e.g. Twilio SID).
	ProviderMessageID ProviderMessageID

	Status

	DestType DestType
}

// State represents the current state of an outgoing message.
type State int

// IsOK returns true if the message has passed successfully to a remote system (StateSending, StateSent, or StateDelivered).
func (s State) IsOK() bool { return s == StateSending || s == StateSent || s == StateDelivered }

const (
	// StateUnknown is returned when the message has not yet been sent.
	StateUnknown State = iota

	// StateSending should be specified when a message is sending but has not been sent.
	// This includes things like remotely queued, ringing, or in-progress calls.
	StateSending

	// StatePending indicates a message waiting to be sent.
	StatePending

	// StateSent means the message has been sent completely, but may not
	// have been delivered (or delivery confirmation is not supported.). For
	// example, an SMS on the carrier network (but not device) or a voice call
	// that rang but got `no-answer`.
	StateSent

	// StateDelivered means the message is completed and was received
	// by the end device. SMS delivery confirmation, or a voice call was
	// completed (including if it was voice mail).
	StateDelivered

	// StateFailedTemp should be set when a message was not sent (no SMS or ringing phone)
	// but a subsequent try later may succeed. (e.g. voice call with busy signal).
	StateFailedTemp

	// StateFailedPerm should be set when a message was not sent (no SMS or ringing phone)
	// but a subsequent attempt will not be expected to succeed. For messages that fail due to
	// invalid config, they should set this state, as without manual intervention, a retry
	// will also fail.
	StateFailedPerm

	// StateBundled indicates that the message has been bundled into another message.
	StateBundled
)

func (s *State) fromString(val string) error {
	switch val {
	case "pending":
		*s = StatePending
	case "sending", "queued_remotely":
		*s = StateSending
	case "sent":
		*s = StateSent
	case "delivered":
		*s = StateDelivered
	case "failed":
		*s = StateFailedPerm
	case "stale":
		*s = StateFailedTemp
	case "bundled":
		*s = StateBundled
	default:
		return fmt.Errorf("unexpected value %q", val)
	}

	return nil
}

func (s *State) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return s.fromString(string(v))
	case string:
		return s.fromString(v)
	case nil:
		*s = StateUnknown
		return nil
	default:
		return fmt.Errorf("unexpected type %T", value)
	}
}
