package notificationchannel

import (
	"database/sql/driver"
	"fmt"

	"github.com/target/goalert/notification"
)

type Type string

const (
	TypeUnknown Type = ""
	TypeSlack   Type = "SLACK"
)

// TypeFromDestType will return the Type associated with a
// notification.DestType.
func TypeFromDestType(t notification.DestType) Type {
	switch t {
	case notification.DestTypeSlackChannel:
		return TypeSlack
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
	case TypeSlack:
		return notification.DestTypeSlackChannel
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
		return fmt.Errorf("could not process unknown type for channel type %T", t)
	}

	return nil
}
