package alertlog

import (
	"database/sql/driver"
	"fmt"
)

// SubjectType represents the type of subject or causer of an alert event.
type SubjectType string

// Possible subject types
const (
	SubjectTypeUser             SubjectType = "user"
	SubjectTypeNoNotification   SubjectType = "no_notification"
	SubjectTypeIntegrationKey   SubjectType = "integration_key"
	SubjectTypeHeartbeatMonitor SubjectType = "heartbeat_monitor"
	SubjectTypeChannel          SubjectType = "channel"
	SubjectTypeNone             SubjectType = ""
)

// Scan handles reading a Type from the DB enum
func (s *SubjectType) Scan(value interface{}) error {
	switch t := value.(type) {
	case []byte:
		*s = SubjectType(t)
	case string:
		*s = SubjectType(t)
	case nil:
		*s = SubjectTypeNone
	default:
		return fmt.Errorf("could not process unknown type %T", t)
	}

	return nil
}
func (s SubjectType) Value() (driver.Value, error) {
	switch s {
	case SubjectTypeUser, SubjectTypeNoNotification, SubjectTypeIntegrationKey, SubjectTypeHeartbeatMonitor, SubjectTypeChannel:
		return string(s), nil
	default:
		return nil, nil
	}
}

// A Subject is generally the causer of an event. If a user closes an alert,
// the entry would have a Subject set to the user.
type Subject struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       SubjectType `json:"type"`
	Classifier string      `json:"classifier"`
}

func subjectString(infinitive bool, s *Subject) string {
	if s == nil {
		return ""
	}
	var str string
	if infinitive {
		str += " to"
	} else {
		switch s.Type {
		case SubjectTypeUser:
			str += " by"
		case SubjectTypeNone:
			return ""
		default:
			str += " via"
		}
	}
	if s.Name == "" {
		str += " [unknown]"
	} else {
		str += " " + s.Name
	}
	switch s.Type {
	case SubjectTypeIntegrationKey:
		str += " integration"
	case SubjectTypeHeartbeatMonitor:
		str += " heartbeat monitor"
	case SubjectTypeNoNotification:
		str += " (no immediate rule)"
	}

	if s.Classifier != "" {
		str += " (" + s.Classifier + ")"
	}
	return str
}
