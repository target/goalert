package graphql2

import (
	"database/sql/driver"
	"fmt"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/schedule"
)

type OnCallNotificationRuleInput struct {
	schedule.OnCallNotificationRule
	Target assignment.RawTarget
}

type MessageLogSegmenyBy string

const (
	SegmentByService     MessageLogSegmenyBy = "service"
	SegmenyByUser        MessageLogSegmenyBy = "user"
	SegmenyByMessageType MessageLogSegmenyBy = "message_type"
)

func (s MessageLogSegmenyBy) Value() (driver.Value, error) {
	str := string(s)
	if str == "" {
		return "", nil
	}
	return str, nil
}

func (s *MessageLogSegmenyBy) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = MessageLogSegmenyBy(t)
	case string:
		*s = MessageLogSegmenyBy(t)
	case nil:
		*s = ""
	default:
		return fmt.Errorf("could not process unknown type for MessageLogSegmenyBy(%T)", t)
	}
	return nil
}
