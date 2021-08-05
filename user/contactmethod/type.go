package contactmethod

import (
	"database/sql/driver"
	"fmt"

	"github.com/target/goalert/notification"
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
)

// Valid returns true if t is a known Type.
func (t Type) Valid() bool {
	return t == TypeVoice || t == TypeSMS || t == TypeEmail || t == TypePush || t == TypeWebhook
}

// TypeFromDestType will return the Type associated with a
// notification.DestType.
func TypeFromDestType(t notification.DestType) Type {
	switch t {
	case notification.DestTypeSMS:
		return TypeSMS
	case notification.DestTypeVoice:
		return TypeVoice
	case notification.DestTypeUserEmail:
		return TypeEmail
	case notification.DestTypeUserWebhook:
		return TypeWebhook
	}

	return TypeUnknown
}

func (t Type) Value() (driver.Value, error) {
	if t == TypeUnknown {
		return nil, nil
	}

	return string(t), nil
}

func (t Type) DestType() notification.DestType {
	switch t {
	case TypeSMS:
		return notification.DestTypeSMS
	case TypeVoice:
		return notification.DestTypeVoice
	case TypeEmail:
		return notification.DestTypeUserEmail
	case TypeWebhook:
		return notification.DestTypeUserWebhook
	}
	return 0
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
