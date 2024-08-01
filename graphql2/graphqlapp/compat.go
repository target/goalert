package graphqlapp

import (
	"context"
	"fmt"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/email"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/user"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// CompatTargetToDest converts an assignment.Target to a gadb.DestV1.
func (a *App) CompatTargetToDest(ctx context.Context, tgt assignment.Target) (gadb.DestV1, error) {
	switch tgt.TargetType() {
	case assignment.TargetTypeUser:
		return gadb.DestV1{
			Type: user.DestTypeUser,
			Args: map[string]string{user.FieldUserID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeRotation:
		return gadb.DestV1{
			Type: rotation.DestTypeRotation,
			Args: map[string]string{rotation.FieldRotationID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeSchedule:
		return gadb.DestV1{
			Type: schedule.DestTypeSchedule,
			Args: map[string]string{schedule.FieldScheduleID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeChanWebhook:
		return gadb.DestV1{
			Type: webhook.DestTypeWebhook,
			Args: map[string]string{webhook.FieldWebhookURL: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeSlackChannel:
		return gadb.DestV1{
			Type: slack.DestTypeSlackChannel,
			Args: map[string]string{slack.FieldSlackChannelID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeNotificationChannel:
		id, err := validate.ParseUUID("TargetID", tgt.TargetID())
		if err != nil {
			return gadb.DestV1{}, err
		}
		dest, err := a.NCStore.FindDestByID(ctx, nil, id)
		if err != nil {
			return gadb.DestV1{}, err
		}

		return dest, nil
	}

	return gadb.DestV1{}, fmt.Errorf("unknown target type: %s", tgt.TargetType())
}

// CompatDestToCMTypeVal converts a gadb.DestV1 to a contactmethod.Type and string value
// for the built-in destination types.
func CompatDestToCMTypeVal(d gadb.DestV1) (graphql2.ContactMethodType, string) {
	switch d.Type {
	case twilio.DestTypeTwilioSMS:
		return graphql2.ContactMethodTypeSms, d.Arg(twilio.FieldPhoneNumber)
	case twilio.DestTypeTwilioVoice:
		return graphql2.ContactMethodTypeVoice, d.Arg(twilio.FieldPhoneNumber)
	case email.DestTypeEmail:
		return graphql2.ContactMethodTypeEmail, d.Arg(email.FieldEmailAddress)
	case webhook.DestTypeWebhook:
		return graphql2.ContactMethodTypeWebhook, d.Arg(webhook.FieldWebhookURL)
	case slack.DestTypeSlackDirectMessage:
		return graphql2.ContactMethodTypeSLACkDm, d.Arg(slack.FieldSlackUserID)
	}

	return "", ""
}

func CompatCMTypeValToDest(cmType graphql2.ContactMethodType, value string) (gadb.DestV1, error) {
	switch cmType {
	case graphql2.ContactMethodTypeEmail:
		return email.NewEmailDest(value), nil
	case graphql2.ContactMethodTypeSms:
		return twilio.NewSMSDest(value), nil
	case graphql2.ContactMethodTypeVoice:
		return twilio.NewVoiceDest(value), nil
	case graphql2.ContactMethodTypeSLACkDm:
		return slack.NewDirectMessageDest(value), nil
	case graphql2.ContactMethodTypeWebhook:
		return webhook.NewWebhookDest(value), nil
	}

	return gadb.DestV1{}, validation.NewFieldError("input.Type", "unsupported type")
}

// CompatDestToTarget converts a gadb.DestV1 to a graphql2.RawTarget
func CompatDestToTarget(d gadb.DestV1) (assignment.RawTarget, error) {
	switch d.Type {
	case user.DestTypeUser:
		return assignment.RawTarget{
			Type: assignment.TargetTypeUser,
			ID:   d.Arg(user.FieldUserID),
		}, nil
	case rotation.DestTypeRotation:
		return assignment.RawTarget{
			Type: assignment.TargetTypeRotation,
			ID:   d.Arg(rotation.FieldRotationID),
		}, nil
	case schedule.DestTypeSchedule:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSchedule,
			ID:   d.Arg(schedule.FieldScheduleID),
		}, nil
	case slack.DestTypeSlackChannel:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackChannel,
			ID:   d.Arg(slack.FieldSlackChannelID),
		}, nil
	case slack.DestTypeSlackUsergroup:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackUserGroup,
			ID:   d.Arg(slack.FieldSlackUsergroupID) + ":" + d.Arg(slack.FieldSlackChannelID),
		}, nil
	case webhook.DestTypeWebhook:
		return assignment.RawTarget{
			Type: assignment.TargetTypeChanWebhook,
			ID:   d.Arg(webhook.FieldWebhookURL),
		}, nil
	}

	return assignment.RawTarget{}, fmt.Errorf("unsupported destination type: %s", d.Type)
}
