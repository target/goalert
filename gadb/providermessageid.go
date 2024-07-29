package gadb

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

var (
	_ driver.Valuer = ProviderMessageID{}
	_ sql.Scanner   = &ProviderMessageID{}
)

// ParseProviderMessageID parses a provider-specific identifier for a message.
func ParseProviderMessageID(id string) (ProviderMessageID, error) {
	var p ProviderMessageID
	err := p.Scan(id)
	return p, err
}

// String returns a parseable string representation of the provider-specific identifier for a message.
func (p ProviderMessageID) String() string {
	if p.ProviderName == "" || p.ExternalID == "" {
		return ""
	}

	return fmt.Sprintf("%s:%s", p.ProviderName, p.ExternalID)
}

func (p ProviderMessageID) Value() (driver.Value, error) {
	val := p.String()
	if val == "" {
		return nil, nil
	}

	return val, nil
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
