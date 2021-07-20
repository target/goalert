package contactmethod

import (
	"fmt"

	"github.com/target/goalert/notification"
)

// Type specifies the medium a ContactMethod is notified.
type Type string

// ContactMethod types
const (
	TypeVoice   Type = "VOICE"
	TypeSMS     Type = "SMS"
	TypeEmail   Type = "EMAIL"
	TypePush    Type = "PUSH"
	TypeWebhook Type = "WEBHOOK"
)

// TypeFromDestType will return the Type associated with a
// notification.DestType.
func TypeFromDestType(t notification.DestType) Type {
	switch t {
	case notification.DestTypeSMS:
		return TypeSMS
	case notification.DestTypeVoice:
		return TypeVoice
	}

	return ""
}

func (t Type) DestType() notification.DestType {
	switch t {
	case TypeSMS:
		return notification.DestTypeSMS
	case TypeVoice:
		return notification.DestTypeVoice
	}
	return 0
}

// Scan handles reading a Type from the DB format
func (r *Type) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*r = Type(t)
	case string:
		*r = Type(t)
	default:
		return fmt.Errorf("could not process unknown type for role %T", t)
	}

	return nil
}
