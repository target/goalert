package alert

import (
	"fmt"
)

// Status is the current state of an Alert.
type Status string

// Alert status types
const (
	StatusTriggered Status = "triggered"
	StatusActive    Status = "active"
	StatusClosed    Status = "closed"
)

func (s Status) Value() string {
	str := string(s)
	if str == "" {
		str = string(StatusTriggered)
	}
	return str
}

func (s *Status) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = Status(t)
	case string:
		*s = Status(t)
	case nil:
		*s = StatusTriggered
	default:
		return fmt.Errorf("could not process unknown type for Status(%T)", t)
	}
	return nil
}
