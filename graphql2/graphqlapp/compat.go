package graphqlapp

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user/contactmethod"
)

// func CompatDestToTarget(d graphql2.DestinationInput) (assignment.RawTarget, error) {
// 	switch d.Type {
// 	case destSlackChan:
// 		return assignment.RawTarget{
// 			Type: assignment.TargetTypeSlackChannel,
// 			ID:   d.FieldValue(fieldSlackChanID),
// 		}, nil
// 	case destSlackUG:
// 		return assignment.RawTarget{
// 			Type: assignment.TargetTypeSlackUserGroup,
// 			ID:   d.FieldValue(fieldSlackUGID) + ":" + d.FieldValue(fieldSlackChanID),
// 		}, nil
// 	case destWebhook:
// 		return assignment.RawTarget{
// 			Type: assignment.TargetTypeChanWebhook,
// 			ID:   d.FieldValue(fieldWebhookURL),
// 		}, nil
// 	}
// }

// CompatDestToCMTypeVal converts a graphql2.DestinationInput to a contactmethod.Type and string value
// for the built-in destination types.
func CompatDestToCMTypeVal(d graphql2.DestinationInput) (contactmethod.Type, string) {
	switch d.Type {
	case destTwilioSMS:
		return contactmethod.TypeSMS, d.FieldValue(fieldPhoneNumber)
	case destTwilioVoice:
		return contactmethod.TypeVoice, d.FieldValue(fieldPhoneNumber)
	case destSMTP:
		return contactmethod.TypeEmail, d.FieldValue(fieldEmailAddress)
	case destWebhook:
		return contactmethod.TypeWebhook, d.FieldValue(fieldWebhookURL)
	case destSlackDM:
		return contactmethod.TypeSlackDM, d.FieldValue(fieldSlackUserID)
	}

	return "", ""
}
func CompatTargetToDest(tgt assignment.Target) (graphql2.Destination, error) {
	switch tgt.TargetType() {
	case assignment.TargetTypeUser:
		return graphql2.Destination{
			Type: destUser,
			Values: []graphql2.FieldValuePair{{
				FieldID: fieldUserID,
				Value:   tgt.TargetID(),
			}}}, nil
	case assignment.TargetTypeRotation:
		return graphql2.Destination{
			Type: destRotation,
			Values: []graphql2.FieldValuePair{{
				FieldID: fieldRotationID,
				Value:   tgt.TargetID(),
			}}}, nil
	case assignment.TargetTypeSchedule:
		return graphql2.Destination{
			Type: destSchedule,
			Values: []graphql2.FieldValuePair{{
				FieldID: fieldScheduleID,
				Value:   tgt.TargetID(),
			}}}, nil
	case assignment.TargetTypeChanWebhook:
		return graphql2.Destination{
			Type: destWebhook,
			Values: []graphql2.FieldValuePair{{
				FieldID: fieldWebhookURL,
				Value:   tgt.TargetID(),
			}}}, nil
	case assignment.TargetTypeSlackChannel:
		return graphql2.Destination{
			Type: destSlackChan,
			Values: []graphql2.FieldValuePair{{
				FieldID: fieldSlackChanID,
				Value:   tgt.TargetID(),
			}}}, nil
	}

	return graphql2.Destination{}, fmt.Errorf("unknown target type: %s", tgt.TargetType())
}
func CompatDestToTarget(d graphql2.DestinationInput) (assignment.RawTarget, error) {
	switch d.Type {
	case destUser:
		return assignment.RawTarget{
			Type: assignment.TargetTypeUser,
			ID:   d.FieldValue(fieldUserID),
		}, nil
	case destRotation:
		return assignment.RawTarget{
			Type: assignment.TargetTypeRotation,
			ID:   d.FieldValue(fieldRotationID),
		}, nil
	case destSchedule:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSchedule,
			ID:   d.FieldValue(fieldScheduleID),
		}, nil
	case destSlackChan:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackChannel,
			ID:   d.FieldValue(fieldSlackChanID),
		}, nil
	case destSlackUG:
		return assignment.RawTarget{
			Type: assignment.TargetTypeSlackUserGroup,
			ID:   d.FieldValue(fieldSlackUGID) + ":" + d.FieldValue(fieldSlackChanID),
		}, nil
	case destWebhook:
		return assignment.RawTarget{
			Type: assignment.TargetTypeChanWebhook,
			ID:   d.FieldValue(fieldWebhookURL),
		}, nil
	}

	return assignment.RawTarget{}, fmt.Errorf("unsupported destination type: %s", d.Type)
}

// func CompatTargetToDest(t assignment.RawTarget) (*graphql2.DestinationInput, error) {
// 	err := validate.OneOf("Target.Type", t.Type, assignment.TargetTypeSlackChannel, assignment.TargetTypeSlackUserGroup, assignment.TargetTypeChanWebhook)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch t.Type {
// 	case assignment.TargetTypeSlackChannel:
// 		return &graphql2.DestinationInput{
// 			Type: destSlackChan,
// 			Values: []graphql2.FieldValueInput{
// 				{
// 					FieldID: fieldSlackChanID,
// 					Value:   t.ID,
// 				},
// 			},
// 		}, nil
// 	case assignment.TargetTypeSlackUserGroup:
// 		ugID, chanID, ok := strings.Cut(t.ID, ":")
// 		if !ok {
// 			return nil, fmt.Errorf("invalid slack usergroup pair: %s", t.ID)
// 		}

// 		return &graphql2.DestinationInput{
// 			Type: destSlackUG,
// 			Values: []graphql2.FieldValueInput{
// 				{
// 					FieldID: fieldSlackUGID,
// 					Value:   ugID,
// 				},
// 				{
// 					FieldID: fieldSlackChanID,
// 					Value:   chanID,
// 				},
// 			},
// 		}, nil

// 	case assignment.TargetTypeChanWebhook:
// 		return &graphql2.DestinationInput{
// 			Type: destWebhook,
// 			Values: []graphql2.FieldValueInput{
// 				{
// 					FieldID: fieldWebhookURL,
// 					Value:   t.ID,
// 				},
// 			},
// 		}, nil
// 	}

// 	// should be unreachable due to validation above
// 	panic("unreachable")
// }

func (a *App) CompatNCToDest(ctx context.Context, ncID uuid.UUID) (*graphql2.Destination, error) {
	nc, err := a.FindOneNC(ctx, ncID)
	if err != nil {
		return nil, err
	}

	switch nc.Type {
	case notificationchannel.TypeSlackChan:
		ch, err := a.SlackStore.Channel(ctx, nc.Value)
		if err != nil {
			return nil, err
		}

		return &graphql2.Destination{
			Type: destSlackChan,
			Values: []graphql2.FieldValuePair{
				{
					FieldID: fieldSlackChanID,
					Value:   nc.Value,
					Label:   ch.Name,
				},
			},
		}, nil
	case notificationchannel.TypeSlackUG:
		ugID, chanID, ok := strings.Cut(nc.Value, ":")
		if !ok {
			return nil, fmt.Errorf("invalid slack usergroup pair: %s", nc.Value)
		}
		ug, err := a.SlackStore.UserGroup(ctx, ugID)
		if err != nil {
			return nil, err
		}
		ch, err := a.SlackStore.Channel(ctx, chanID)
		if err != nil {
			return nil, err
		}

		return &graphql2.Destination{
			Type: destSlackUG,
			Values: []graphql2.FieldValuePair{
				{
					FieldID: fieldSlackUGID,
					Value:   ugID,
					Label:   ug.Handle,
				},
				{
					FieldID: fieldSlackChanID,
					Value:   chanID,
					Label:   ch.Name,
				},
			},
		}, nil
	case notificationchannel.TypeWebhook:
		u, err := url.Parse(nc.Value)
		if err != nil {
			return nil, err
		}

		return &graphql2.Destination{
			Type: destWebhook,
			Values: []graphql2.FieldValuePair{
				{
					FieldID: fieldWebhookURL,
					Value:   nc.Value,
					Label:   u.Hostname(),
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported notification channel type: %s", nc.Type)
	}
}
