package notification

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
)

// ProviderMessageID is a provider-specific identifier for a message.
type ProviderMessageID struct {
	// ExternalID is the provider-specific identifier for the message.
	ExternalID   string
	ProviderName string
}

var _ driver.Valuer = ProviderMessageID{}
var _ sql.Scanner = &ProviderMessageID{}

func (p ProviderMessageID) Value() (driver.Value, error) {
	if p.ProviderName == "" || p.ExternalID == "" {
		return nil, nil
	}

	return p.ProviderName + ":" + p.ExternalID, nil
}
func (p *ProviderMessageID) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		if !strings.Contains(v, ":") {
			return fmt.Errorf("invalid provider id format: '%s'; expected 'providername:providerid'", v)
		}
		parts := strings.SplitN(v, ":", 2)
		p.ProviderName = parts[0]
		p.ExternalID = parts[1]
	case nil:
		p.ExternalID = ""
		p.ProviderName = ""
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}

	return nil
}
