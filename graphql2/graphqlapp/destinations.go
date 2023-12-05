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

func (q *Query) InputFieldValidate(ctx context.Context, dataType, value string) (bool, error) {
	return false, nil
}

func (q *Query) InputFieldSearch(ctx context.Context, dataType string, input graphql2.InputFieldSearchInput) (conn *graphql2.DestinationInfoConnection, err error) {
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
				RequiredFields: []graphql2.InputFieldConfig{{
					DataType:           "PHONE",
					LabelSingular:      "Phone Number",
					LabelPlural:        "Phone Numbers",
					Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
					PlaceholderText:    "11235550123",
					UserDisclaimer:     cfg.General.NotificationDisclaimer,
					Prefix:             "+",
					InputType:          "tel",
					SupportsValidation: true,
				}},
			},
			{
				TypeID:          "VOICE",
				Name:            "Voice Call",
				Enabled:         cfg.Twilio.Enable,
				DisabledMessage: "Twilio must be configured by an administrator",
				RequiredFields: []graphql2.InputFieldConfig{{
					DataType:           "PHONE",
					LabelSingular:      "Phone Number",
					LabelPlural:        "Phone Numbers",
					Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
					PlaceholderText:    "11235550123",
					UserDisclaimer:     cfg.General.NotificationDisclaimer,
					Prefix:             "+",
					InputType:          "tel",
					SupportsValidation: true,
				}},
			},
			{
				TypeID:          "EMAIL",
				Name:            "Email",
				Enabled:         cfg.SMTP.Enable,
				DisabledMessage: "SMTP must be configured by an administrator",
				RequiredFields: []graphql2.InputFieldConfig{{
					DataType:           "EMAIL",
					LabelSingular:      "Email Address",
					LabelPlural:        "Email Addresses",
					PlaceholderText:    "foobar@example.com",
					InputType:          "email",
					SupportsValidation: true,
				}},
			},
			{
				TypeID:          "WEBHOOK",
				Name:            "Webhook",
				Enabled:         cfg.Webhook.Enable,
				DisabledMessage: "Webhooks must be enabled by an administrator",
				RequiredFields: []graphql2.InputFieldConfig{{
					DataType:           "URL",
					LabelSingular:      "URL",
					LabelPlural:        "URLs",
					PlaceholderText:    "https://example.com",
					InputType:          "url",
					SupportsValidation: true,
				}},
			},
			{
				TypeID:          "SLACK_DM",
				Name:            "Slack DM",
				Enabled:         cfg.Slack.Enable,
				DisabledMessage: "Slack must be enabled by an administrator",
				RequiredFields: []graphql2.InputFieldConfig{{
					DataType:           "SLACK_USER_ID",
					LabelSingular:      "Slack User",
					LabelPlural:        "Slack Users",
					PlaceholderText:    "member ID",
					InputType:          "text",
					SupportsValidation: true,
					IsSearchSelectable: true,
				}},
			},
		}, nil
	}

	return nil, nil
}
