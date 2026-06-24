package nfymsg

import "time"

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

	// Age is the time since the message was first sent. Ignored for new messages.
	Age time.Duration
}
