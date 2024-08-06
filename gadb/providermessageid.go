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
	// Older versions of GoAlert had a separate name for each provider from the destination type, so we need to map them for compatibility.
	//
	// Since the SMS and voice message types are the only ones that rely on async status updates, they are the only ones that require this mapping.
	switch p.ProviderName {
	case "builtin-twilio-sms":
		p.ProviderName = "Twilio-SMS"
	case "builtin-twilio-voice":
		p.ProviderName = "Twilio-Voice"
	}
	val := p.String()
	if val == "" {
		return nil, nil
	}

	return val, nil
}

func (p *ProviderMessageID) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		var ok bool
		p.ProviderName, p.ExternalID, ok = strings.Cut(v, ":")
		if !ok {
			return fmt.Errorf("invalid provider id format: '%s'; expected 'providername:providerid'", v)
		}

		// Older versions of GoAlert had a separate name for each provider from the destination type, so we need to map them for compatibility.
		//
		// Since the SMS and voice message types are the only ones that rely on async status updates, they are the only ones that require this mapping.
		switch p.ProviderName {
		case "Twilio-SMS":
			p.ProviderName = "builtin-twilio-sms"
		case "Twilio-Voice":
			p.ProviderName = "builtin-twilio-voice"
		}
	case nil:
		p.ExternalID = ""
		p.ProviderName = ""
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}

	return nil
}
