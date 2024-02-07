package graphqlapp

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/user/contactmethod"
)

// CompatNCToDest converts a notification channel to a destination.
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
