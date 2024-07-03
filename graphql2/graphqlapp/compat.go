package graphqlapp

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user/contactmethod"
)

// CompatTargetToDest converts an assignment.Target to a graphql2.Destination.
func CompatTargetToDest(tgt assignment.Target) (graphql2.Destination, error) {
	switch tgt.TargetType() {
	case assignment.TargetTypeUser:
		return graphql2.Destination{
			Type: destUser,
			Args: map[string]string{fieldUserID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeRotation:
		return graphql2.Destination{
			Type: destRotation,
			Args: map[string]string{fieldRotationID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeSchedule:
		return graphql2.Destination{
			Type: destSchedule,
			Args: map[string]string{fieldScheduleID: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeChanWebhook:
		return graphql2.Destination{
			Type: destWebhook,
			Args: map[string]string{fieldWebhookURL: tgt.TargetID()},
		}, nil
	case assignment.TargetTypeSlackChannel:
		return graphql2.Destination{
			Type: destSlackChan,
			Args: map[string]string{fieldSlackChanID: tgt.TargetID()},
		}, nil
	}

	return graphql2.Destination{}, fmt.Errorf("unknown target type: %s", tgt.TargetType())
}

// CompatNCToDest converts a notification channel to a destination.
func (a *App) CompatNCToDest(ctx context.Context, ncID uuid.UUID) (*graphql2.Destination, error) {
	nc, err := a.FindOneNC(ctx, ncID)
	if err != nil {
		return nil, err
	}

	switch nc.Type {
	case notificationchannel.TypeSlackChan:
		return &graphql2.Destination{
			Type: destSlackChan,
			Args: map[string]string{fieldSlackChanID: nc.Value},
		}, nil
	case notificationchannel.TypeSlackUG:
		ugID, chanID, ok := strings.Cut(nc.Value, ":")
		if !ok {
			return nil, fmt.Errorf("invalid slack usergroup pair: %s", nc.Value)
		}

		return &graphql2.Destination{
			Type: destSlackUG,
			Args: map[string]string{
				fieldSlackUGID:   ugID,
				fieldSlackChanID: chanID,
			},
		}, nil
	case notificationchannel.TypeWebhook:
		return &graphql2.Destination{
			Type: destWebhook,
			Args: map[string]string{fieldWebhookURL: nc.Value},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported notification channel type: %s", nc.Type)
	}
}

// CompatDestToCMTypeVal converts a graphql2.DestinationInput to a contactmethod.Type and string value
// for the built-in destination types.
func CompatDestToCMTypeVal(d graphql2.DestinationInput) (contactmethod.Type, string) {
	switch d.Type {
	case destTwilioSMS:
		return contactmethod.TypeSMS, d.Args[fieldPhoneNumber]
	case destTwilioVoice:
		return contactmethod.TypeVoice, d.Args[fieldPhoneNumber]
	case destSMTP:
		return contactmethod.TypeEmail, d.Args[fieldEmailAddress]
	case destWebhook:
		return contactmethod.TypeWebhook, d.Args[fieldWebhookURL]
	case destSlackDM:
		return contactmethod.TypeSlackDM, d.Args[fieldSlackUserID]
	}

	return "", ""
}

// CompatDestToTarget converts a graphql2.DestinationInput to a graphql2.RawTarget
func CompatDestToTarget(d graphql2.DestinationInput) (assignment.RawTarget, error) {
	switch d.Type {
	case destUser:
		return assignment.RawTarget{
			Type: assignment.TargetTypeUser,
			ID:   d.Args[fieldUserID],
		}, nil
	case destRotation:
		return assignment.RawTarget{
			Type: assignment.TargetTypeRotation,
			ID:   d.Args[fieldRotationID],
		}, nil
	case destSchedule:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSchedule,
			ID:   d.Args[fieldScheduleID],
		}, nil
	case destSlackChan:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackChannel,
			ID:   d.Args[fieldSlackChanID],
		}, nil
	case destSlackUG:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackUserGroup,
			ID:   d.Args[fieldSlackUGID] + ":" + d.Args[fieldSlackChanID],
		}, nil
	case destWebhook:
		return assignment.RawTarget{
			Type: assignment.TargetTypeChanWebhook,
			ID:   d.Args[fieldWebhookURL],
		}, nil
	}

	return assignment.RawTarget{}, fmt.Errorf("unsupported destination type: %s", d.Type)
}
