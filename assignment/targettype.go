package assignment

//go:generate go tool stringer -type TargetType

import (
	"encoding"
	"io"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/validation"
)

// TargetType represents the destination type of an assignment
type TargetType int

// Assignment destination types
const (
	TargetTypeUnspecified TargetType = iota
	TargetTypeEscalationPolicy
	TargetTypeNotificationPolicy
	TargetTypeRotation
	TargetTypeService
	TargetTypeSchedule
	TargetTypeCalendarSubscription
	TargetTypeUser
	TargetTypeNotificationChannel
	TargetTypeSlackChannel
	TargetTypeSlackUserGroup
	TargetTypeChanWebhook
	TargetTypeIntegrationKey
	TargetTypeUserOverride
	TargetTypeNotificationRule
	TargetTypeContactMethod
	TargetTypeHeartbeatMonitor
	TargetTypeUserSession
)

var (
	_ graphql.Marshaler        = TargetType(0)
	_ graphql.Unmarshaler      = new(TargetType)
	_ encoding.TextMarshaler   = TargetType(0)
	_ encoding.TextUnmarshaler = new(TargetType)
)

func (tt *TargetType) UnmarshalText(data []byte) error {
	str := string(data)
	switch str {
	case "escalationPolicy":
		*tt = TargetTypeEscalationPolicy
	case "notificationPolicy":
		*tt = TargetTypeNotificationPolicy
	case "rotation":
		*tt = TargetTypeRotation
	case "service":
		*tt = TargetTypeService
	case "schedule":
		*tt = TargetTypeSchedule
	case "calendarSubscription":
		*tt = TargetTypeCalendarSubscription
	case "user":
		*tt = TargetTypeUser
	case "integrationKey":
		*tt = TargetTypeIntegrationKey
	case "notificationChannel":
		*tt = TargetTypeNotificationChannel
	case "slackChannel":
		*tt = TargetTypeSlackChannel
	case "slackUserGroup":
		*tt = TargetTypeSlackUserGroup
	case "chanWebhook":
		*tt = TargetTypeChanWebhook
	case "userOverride":
		*tt = TargetTypeUserOverride
	case "contactMethod":
		*tt = TargetTypeContactMethod
	case "notificationRule":
		*tt = TargetTypeNotificationRule
	case "heartbeatMonitor":
		*tt = TargetTypeHeartbeatMonitor
	case "userSession":
		*tt = TargetTypeUserSession
	default:
		return validation.NewFieldError("TargetType", "unknown target type "+str)
	}

	return nil
}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (tt *TargetType) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}
	return tt.UnmarshalText([]byte(str))
}

func (tt TargetType) MarshalText() ([]byte, error) {
	switch tt {
	case TargetTypeEscalationPolicy:
		return []byte("escalationPolicy"), nil
	case TargetTypeNotificationPolicy:
		return []byte("notificationPolicy"), nil
	case TargetTypeRotation:
		return []byte("rotation"), nil
	case TargetTypeService:
		return []byte("service"), nil
	case TargetTypeSchedule:
		return []byte("schedule"), nil
	case TargetTypeCalendarSubscription:
		return []byte("calendarSubscription"), nil
	case TargetTypeUser:
		return []byte("user"), nil
	case TargetTypeIntegrationKey:
		return []byte("integrationKey"), nil
	case TargetTypeUserOverride:
		return []byte("userOverride"), nil
	case TargetTypeNotificationChannel:
		return []byte("notificationChannel"), nil
	case TargetTypeSlackChannel:
		return []byte("slackChannel"), nil
	case TargetTypeSlackUserGroup:
		return []byte("slackUserGroup"), nil
	case TargetTypeChanWebhook:
		return []byte("chanWebhook"), nil
	case TargetTypeContactMethod:
		return []byte("contactMethod"), nil
	case TargetTypeNotificationRule:
		return []byte("notificationRule"), nil
	case TargetTypeHeartbeatMonitor:
		return []byte("heartbeatMonitor"), nil
	case TargetTypeUserSession:
		return []byte("userSession"), nil
	}

	return nil, validation.NewFieldError("TargetType", "unknown target type "+tt.String())
}

// MarshalGQL implements the graphql.Marshaler interface
func (tt TargetType) MarshalGQL(w io.Writer) {
	data, err := tt.MarshalText()
	if err != nil {
		panic(err)
	}
	graphql.MarshalString(string(data)).MarshalGQL(w)
}
