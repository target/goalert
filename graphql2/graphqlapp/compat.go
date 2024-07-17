package graphqlapp

import (
	"fmt"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rotation"
	"github.com/target/goalert/user"
	"github.com/target/goalert/user/contactmethod"
)

// CompatTargetToDest converts an assignment.Target to a gadb.DestV1.
func CompatTargetToDest(tgt assignment.Target) (gadb.DestV1, error) {
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
	}

	return gadb.DestV1{}, fmt.Errorf("unknown target type: %s", tgt.TargetType())
}

// CompatDestToCMTypeVal converts a gadb.DestV1 to a contactmethod.Type and string value
// for the built-in destination types.
func CompatDestToCMTypeVal(d gadb.DestV1) (contactmethod.Type, string) {
	switch d.Type {
	case destTwilioSMS:
		return contactmethod.TypeSMS, d.Arg(fieldPhoneNumber)
	case destTwilioVoice:
		return contactmethod.TypeVoice, d.Arg(fieldPhoneNumber)
	case destSMTP:
		return contactmethod.TypeEmail, d.Arg(fieldEmailAddress)
	case webhook.DestTypeWebhook:
		return contactmethod.TypeWebhook, d.Arg(webhook.FieldWebhookURL)
	case slack.DestTypeSlackDirectMessage:
		return contactmethod.TypeSlackDM, d.Arg(slack.FieldSlackUserID)
	}

	return "", ""
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
