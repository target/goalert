package graphqlapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user/contactmethod"
)

// CompatTargetToDest converts an assignment.Target to a gadb.DestV1.
func CompatTargetToDest(tgt assignment.Target) (gadb.DestV1, error) {
	switch tgt.TargetType() {
	case assignment.TargetTypeUser:
		return gadb.DestV1{
			Type: destUser,
			Args: map[string]string{fieldUserID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeRotation:
		return gadb.DestV1{
			Type: destRotation,
			Args: map[string]string{fieldRotationID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeSchedule:
		return gadb.DestV1{
			Type: destSchedule,
			Args: map[string]string{fieldScheduleID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeChanWebhook:
		return gadb.DestV1{
			Type: destWebhook,
			Args: map[string]string{fieldWebhookURL: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeSlackChannel:
		return gadb.DestV1{
			Type: slack.DestTypeSlackChannel,
			Args: map[string]string{slack.FieldSlackChannelID: tgt.TargetID()},
		}, nil
	}

	return gadb.DestV1{}, fmt.Errorf("unknown target type: %s", tgt.TargetType())
}

// CompatNCToDest converts a notification channel to a destination.
func (a *App) CompatNCToDest(ctx context.Context, ncID uuid.UUID) (*gadb.DestV1, error) {
	nc, err := a.FindOneNC(ctx, ncID)
	if err != nil {
		return nil, err
	}

	switch nc.Type {
	case notificationchannel.TypeSlackChan:
		return &gadb.DestV1{
			Type: slack.DestTypeSlackChannel,
			Args: map[string]string{slack.FieldSlackChannelID: nc.Value},
		}, nil
	case notificationchannel.TypeSlackUG:
		ugID, chanID, ok := strings.Cut(nc.Value, ":")
		if !ok {
			return nil, fmt.Errorf("invalid slack usergroup pair: %s", nc.Value)
		}

		return &gadb.DestV1{
			Type: destSlackUG,
			Args: map[string]string{
				fieldSlackUGID:            ugID,
				slack.FieldSlackChannelID: chanID,
			},
		}, nil
	case notificationchannel.TypeWebhook:
		return &gadb.DestV1{
			Type: destWebhook,
			Args: map[string]string{fieldWebhookURL: nc.Value},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported notification channel type: %s", nc.Type)
	}
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
	case destWebhook:
		return contactmethod.TypeWebhook, d.Arg(fieldWebhookURL)
	case destSlackDM:
		return contactmethod.TypeSlackDM, d.Arg(fieldSlackUserID)
	}

	return "", ""
}

// CompatDestToTarget converts a gadb.DestV1 to a graphql2.RawTarget
func CompatDestToTarget(d gadb.DestV1) (assignment.RawTarget, error) {
	switch d.Type {
	case destUser:
		return assignment.RawTarget{
			Type: assignment.TargetTypeUser,
			ID:   d.Arg(fieldUserID),
		}, nil
	case destRotation:
		return assignment.RawTarget{
			Type: assignment.TargetTypeRotation,
			ID:   d.Arg(fieldRotationID),
		}, nil
	case destSchedule:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSchedule,
			ID:   d.Arg(fieldScheduleID),
		}, nil
	case slack.DestTypeSlackChannel:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackChannel,
			ID:   d.Arg(slack.FieldSlackChannelID),
		}, nil
	case destSlackUG:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackUserGroup,
			ID:   d.Arg(fieldSlackUGID) + ":" + d.Arg(slack.FieldSlackChannelID),
		}, nil
	case destWebhook:
		return assignment.RawTarget{
			Type: assignment.TargetTypeChanWebhook,
			ID:   d.Arg(fieldWebhookURL),
		}, nil
	}

	return assignment.RawTarget{}, fmt.Errorf("unsupported destination type: %s", d.Type)
}
