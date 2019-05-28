package assignment

//go:generate go run golang.org/x/tools/cmd/stringer -type TargetType

import (
	"github.com/target/goalert/validation"
	"io"

	"github.com/99designs/gqlgen/graphql"
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
	TargetTypeUser
	TargetTypeNotificationChannel
	TargetTypeSlackChannel
	TargetTypeIntegrationKey
	TargetTypeUserOverride
	TargetTypeNotificationRule
	TargetTypeContactMethod
)

// UnmarshalGQL implements the graphql.Marshaler interface
func (tt *TargetType) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err != nil {
		return err
	}

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
	case "user":
		*tt = TargetTypeUser
	case "integrationKey":
		*tt = TargetTypeIntegrationKey
	case "notificationChannel":
		*tt = TargetTypeNotificationChannel
	case "slackChannel":
		*tt = TargetTypeSlackChannel
	case "userOverride":
		*tt = TargetTypeUserOverride
	case "contactMethod":
		*tt = TargetTypeContactMethod
	case "notificationRule":
		*tt = TargetTypeNotificationRule
	default:
		return validation.NewFieldError("TargetType", "unknown target type "+str)
	}

	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (tt TargetType) MarshalGQL(w io.Writer) {
	switch tt {
	case TargetTypeEscalationPolicy:
		graphql.MarshalString("escalationPolicy").MarshalGQL(w)
	case TargetTypeNotificationPolicy:
		graphql.MarshalString("notificationPolicy").MarshalGQL(w)
	case TargetTypeRotation:
		graphql.MarshalString("rotation").MarshalGQL(w)
	case TargetTypeService:
		graphql.MarshalString("service").MarshalGQL(w)
	case TargetTypeSchedule:
		graphql.MarshalString("schedule").MarshalGQL(w)
	case TargetTypeUser:
		graphql.MarshalString("user").MarshalGQL(w)
	case TargetTypeIntegrationKey:
		graphql.MarshalString("integrationKey").MarshalGQL(w)
	case TargetTypeUserOverride:
		graphql.MarshalString("userOverride").MarshalGQL(w)
	case TargetTypeNotificationChannel:
		graphql.MarshalString("notificationChannel").MarshalGQL(w)
	case TargetTypeSlackChannel:
		graphql.MarshalString("slackChannel").MarshalGQL(w)
	case TargetTypeContactMethod:
		graphql.MarshalString("contactMethod").MarshalGQL(w)
	case TargetTypeNotificationRule:
		graphql.MarshalString("notificationRule").MarshalGQL(w)
	}
}
