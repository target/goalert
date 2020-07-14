package notification

import "context"

// MessageStatus represents the state of an outgoing message.
type MessageStatus struct {
	// Ctx is the context of this status update (used for tracing if provided).
	Ctx context.Context

	// ID is the GoAlert message ID.
	ID string

	// ProviderMessageID is a string that represents the provider-specific ID of the message (e.g. Twilio SID).
	ProviderMessageID string

	// State is the current state.
	State MessageState

	// Details can contain any additional information about the State (e.g. "ringing", "no-answer" etc..).
	Details string

	//LastStatus labels information given in details for quick overall (failed, delivered, etc..).
	LastStatus MessageLastStatus

	// Sequence can be used when the provider sends updates out-of order (e.g. Twilio).
	// The Sequence number defaults to 0, and a status update is ignored unless it's
	// Sequence number is >= the current one.
	Sequence int
}

func (stat *MessageStatus) wrap(ctx context.Context, n *namedSender) *MessageStatus {
	if stat == nil {
		return nil
	}

	s := *stat
	if ctx != nil {
		s.Ctx = ctx
	}
	s.ProviderMessageID = n.name + ":" + s.ProviderMessageID
	return &s
}

// MessageState represents the current state of an outgoing message.
type MessageState int

const (
	// MessageStateSending should be specified when a message is sending but has not been sent.
	// This includes things like ringing, or in-progress calls.
	MessageStateSending MessageState = iota

	//MessageStatePending includes things like remotely queued.
	MessageStatePending

	// MessageStateSent means the message has been sent completely, but may not
	// have been delivered (or delivery confirmation is not supported.). For
	// example, an SMS on the carrier network (but not device) or a voice call
	// that rang but got `no-answer`.
	MessageStateSent

	// MessageStateDelivered means the message is completed and was received
	// by the end device. SMS delivery confirmation, or a voice call was
	// completed (including if it was voice mail).
	MessageStateDelivered

	// MessageStateFailedTemp should be set when a message was not sent (no SMS or ringing phone)
	// but a subsequent try later may succeed. (e.g. voice call with busy signal).
	MessageStateFailedTemp

	// MessageStateFailedPerm should be set when a message was not sent (no SMS or ringing phone)
	// but a subsequent attempt will not be expected to succeed. For messages that fail due to
	// invalid config, they should set this state, as without manual intervention, a retry
	// will also fail.
	MessageStateFailedPerm
)

type MessageLastStatus string

const (
	MessageLastStatusPending        MessageLastStatus = "pending"
	MessageLastStatusSending        MessageLastStatus = "sending"
	MessageLastStatusQueuedRemotely MessageLastStatus = "queued_remotely"
	MessageLastStatusSent           MessageLastStatus = "sent"
	MessageLastStatusDelivered      MessageLastStatus = "delivered"
	MessageLastStatusFailed         MessageLastStatus = "failed"
	MessageLastStatusBundled        MessageLastStatus = "bundled"
)
