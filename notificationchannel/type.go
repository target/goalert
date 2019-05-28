package notificationchannel

import (
	"fmt"
	"github.com/target/goalert/notification"
)

type Type string

const (
	TypeSlack Type = "SLACK"
)

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
	case []byte:
		*r = Type(t)
	case string:
		*r = Type(t)
	default:
		return fmt.Errorf("could not process unknown type for channel type %T", t)
	}

	return nil
}
