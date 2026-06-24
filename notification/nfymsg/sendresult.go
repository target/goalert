package nfymsg

import "github.com/target/goalert/gadb"

// SendResult represents the result of a sent message.
type SendResult struct {
	// ID is the GoAlert message ID.
	ID string

	// ProviderMessageID is an identifier that represents the provider-specific ID
	// of the message (e.g. Twilio SID).
	ProviderMessageID gadb.ProviderMessageID

	Status

	DestType string
}
