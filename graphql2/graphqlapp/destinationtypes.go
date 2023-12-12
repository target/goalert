package graphqlapp

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
)

func (q *Query) DestinationTypes(ctx context.Context) ([]graphql2.DestinationTypeInfo, error) {
	cfg := config.FromContext(ctx)
	return []graphql2.DestinationTypeInfo{
		{
			Type:            "SMS",
			Name:            "Text Message (SMS)",
			Enabled:         cfg.Twilio.Enable,
			DisabledMessage: "Twilio must be configured by an administrator",
			UserDisclaimer:  cfg.General.NotificationDisclaimer,
			IsContactMethod: true,
			RequiredFields: []graphql2.InputFieldConfig{{
				DataType:           "PHONE",
				LabelSingular:      "Phone Number",
				LabelPlural:        "Phone Numbers",
				Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
				PlaceholderText:    "11235550123",
				Prefix:             "+",
				InputType:          "tel",
				SupportsValidation: true,
			}},
		},
		{
			Type:            "VOICE",
			Name:            "Voice Call",
			Enabled:         cfg.Twilio.Enable,
			DisabledMessage: "Twilio must be configured by an administrator",
			UserDisclaimer:  cfg.General.NotificationDisclaimer,
			IsContactMethod: true,
			RequiredFields: []graphql2.InputFieldConfig{{
				DataType:           "PHONE",
				LabelSingular:      "Phone Number",
				LabelPlural:        "Phone Numbers",
				Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
				PlaceholderText:    "11235550123",
				Prefix:             "+",
				InputType:          "tel",
				SupportsValidation: true,
			}},
		},
		{
			Type:            "EMAIL",
			Name:            "Email",
			Enabled:         cfg.SMTP.Enable,
			IsContactMethod: true,
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
			Type:            "WEBHOOK",
			Name:            "Webhook",
			Enabled:         cfg.Webhook.Enable,
			IsContactMethod: true,
			IsEPTarget:      true,
			DisabledMessage: "Webhooks must be enabled by an administrator",
			RequiredFields: []graphql2.InputFieldConfig{{
				DataType:           "URL",
				LabelSingular:      "Webhook URL",
				LabelPlural:        "Webhook URLs",
				PlaceholderText:    "https://example.com",
				InputType:          "url",
				Hint:               "Webhook Documentation",
				HintURL:            "/docs#webhooks",
				SupportsValidation: true,
			}},
		},
		{
			Type:            "SLACK_DM",
			Name:            "Slack Message (DM)",
			Enabled:         cfg.Slack.Enable,
			IsContactMethod: true,
			DisabledMessage: "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.InputFieldConfig{{
				DataType:        "SLACK_USER_ID",
				LabelSingular:   "Slack User",
				LabelPlural:     "Slack Users",
				PlaceholderText: "member ID",
				InputType:       "text",
				// IsSearchSelectable: true, // TODO: implement search select functionality for users
				Hint: `Go to your Slack profile, click the three dots, and select "Copy member ID".`,
			}},
		},
	}, nil
}
