package graphqlapp

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
)

// TODO: switch TypeID to custom type
// TODO: sort destination types (i.e., disabled last)

func (q *Query) Destination(ctx context.Context, typeID, value string) (*graphql2.DestinationInfo, error) {
	return nil, nil
}

func (q *Query) DestinationInputIsValid(ctx context.Context, typeID, value string) (bool, error) {
	return false, nil
}

func (q *Query) DestinationSearch(ctx context.Context, typeID string, input graphql2.DestinationSearchInput) (conn *graphql2.DestinationInfoConnection, err error) {
	return nil, nil
}

func (q *Query) DestinationType(ctx context.Context, typeID string) (*graphql2.DestinationTypeInfo, error) {
	return nil, nil
}

func (q *Query) DestinationTypes(ctx context.Context, isContactMethod, isEPTarget, isSchedOnCallNotify *bool) ([]graphql2.DestinationTypeInfo, error) {
	cfg := config.FromContext(ctx)
	if isContactMethod != nil && *isContactMethod {
		return []graphql2.DestinationTypeInfo{
			{
				TypeID:          "SMS",
				Name:            "Text Message (SMS)",
				Enabled:         cfg.Twilio.Enable,
				DisabledMessage: "Twilio must be configured by an administrator",
				Input: &graphql2.DestinationInput{
					TypeID:             "SMS",
					NameSingular:       "Phone Number",
					NamePlural:         "Phone Numbers",
					Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
					PlaceholderText:    "11235550123",
					UserDisclaimer:     cfg.General.NotificationDisclaimer,
					Prefix:             "+",
					InputType:          "tel",
					SupportsValidation: true,
				},
			},
			{
				TypeID:          "VOICE",
				Name:            "Voice Call",
				Enabled:         cfg.Twilio.Enable,
				DisabledMessage: "Twilio must be configured by an administrator",
				Input: &graphql2.DestinationInput{
					TypeID:             "VOICE",
					NameSingular:       "Phone Number",
					NamePlural:         "Phone Numbers",
					Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
					PlaceholderText:    "11235550123",
					UserDisclaimer:     cfg.General.NotificationDisclaimer,
					Prefix:             "+",
					InputType:          "tel",
					SupportsValidation: true,
				},
			},
			{
				TypeID:          "EMAIL",
				Name:            "Email",
				Enabled:         cfg.SMTP.Enable,
				DisabledMessage: "SMTP must be configured by an administrator",
				Input: &graphql2.DestinationInput{
					TypeID:             "EMAIL",
					NameSingular:       "Email Address",
					NamePlural:         "Email Addresses",
					PlaceholderText:    "foobar@example.com",
					InputType:          "email",
					SupportsValidation: true,
				},
			},
			{
				TypeID:          "WEBHOOK",
				Name:            "Webhook",
				Enabled:         cfg.Webhook.Enable,
				DisabledMessage: "Webhooks must be enabled by an administrator",
				Input: &graphql2.DestinationInput{
					TypeID:             "WEBHOOK",
					NameSingular:       "URL",
					NamePlural:         "URLs",
					PlaceholderText:    "https://example.com",
					InputType:          "url",
					SupportsValidation: true,
				},
			},
			{
				TypeID:          "SLACK_DM",
				Name:            "Slack DM",
				Enabled:         cfg.Slack.Enable,
				DisabledMessage: "Slack must be enabled by an administrator",
				Input: &graphql2.DestinationInput{
					TypeID:             "SLACK_DM",
					NameSingular:       "Slack User",
					NamePlural:         "Slack Users",
					PlaceholderText:    "member ID",
					InputType:          "text",
					SupportsValidation: true,
					IsSearchSelectable: true,
				},
			},
		}, nil
	}

	return nil, nil
}
