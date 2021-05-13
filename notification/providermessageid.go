package notification

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
)

// ProviderMessageID is a provider-specific identifier for a message.
type ProviderMessageID struct {
	ID       string
	Provider string
}

var _ driver.Valuer = ProviderMessageID{}
var _ sql.Scanner = &ProviderMessageID{}

func (p ProviderMessageID) Value() (driver.Value, error) {
	if p.Provider == "" || p.ID == "" {
		return nil, nil
	}

	return p.Provider + ":" + p.ID, nil
}
func (p *ProviderMessageID) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		if !strings.Contains(v, ":") {
			return fmt.Errorf("invalid provider id format: '%s'; expected 'providername:providerid'", v)
		}
		parts := strings.SplitN(v, ":", 2)
		p.Provider = parts[0]
		p.ID = parts[1]
	case nil:
		p.ID = ""
		p.Provider = ""
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}

	return nil
}
