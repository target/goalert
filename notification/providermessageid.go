package notification

import "github.com/target/goalert/gadb"

type ProviderMessageID = gadb.ProviderMessageID

// ParseProviderMessageID parses a provider-specific identifier for a message.
func ParseProviderMessageID(id string) (ProviderMessageID, error) {
	var p ProviderMessageID
	err := p.Scan(id)
	return p, err
}
