package contactmethod

import (
	"database/sql/driver"
	"fmt"
)

// Type specifies the medium a ContactMethod is notified.
type Type string

// ContactMethod types
const (
	TypeUnknown Type = ""
	TypeVoice   Type = "VOICE"
	TypeSMS     Type = "SMS"
	TypeEmail   Type = "EMAIL"
	TypePush    Type = "PUSH"
	TypeWebhook Type = "WEBHOOK"
	TypeSlackDM Type = "SLACK_DM"
)

func (t Type) StatusUpdatesAlways() bool {
	return t == TypeSlackDM || t == TypeWebhook
}

func (t Type) StatusUpdatesNever() bool {
	return t == TypePush
}

func (t Type) Value() (driver.Value, error) {
	if t == TypeUnknown {
		return nil, nil
	}

	return string(t), nil
}

// Scan handles reading a Type from the DB format
func (r *Type) Scan(value interface{}) error {
	switch t := value.(type) {
	case nil:
		*r = TypeUnknown
	case []byte:
		*r = Type(t)
	case string:
		*r = Type(t)
	default:
		return fmt.Errorf("could not process unknown type for role %T", t)
	}

	return nil
}
