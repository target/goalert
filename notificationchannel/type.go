package notificationchannel

import (
	"database/sql/driver"
	"fmt"
)

type Type string

const (
	TypeUnknown   Type = ""
	TypeSlackChan Type = "SLACK"
	TypeWebhook   Type = "WEBHOOK"
	TypeSlackUG   Type = "SLACK_USER_GROUP"
)

// Valid returns true if t is a known Type.
func (t Type) Valid() bool {
	return t == TypeSlackChan
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
		return fmt.Errorf("could not process unknown type for channel type %T", t)
	}

	return nil
}
