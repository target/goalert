package graphqlapp

import (
	"context"
	"slices"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// builtin-types
const (
	destTwilioSMS   = "builtin-twilio-sms"
	destTwilioVoice = "builtin-twilio-voice"
	destSMTP        = "builtin-smtp-email"
	destWebhook     = "builtin-webhook"
	destSlackDM     = "builtin-slack-dm"
	destSlackChan   = "builtin-slack-channel"
	destSlackUG     = "builtin-slack-usergroup"
	destUser        = "builtin-user"
	destRotation    = "builtin-rotation"
	destSchedule    = "builtin-schedule"

	fieldPhoneNumber  = "phone-number"
	fieldEmailAddress = "email-address"
	fieldWebhookURL   = "webhook-url"
	fieldSlackUserID  = "slack-user-id"
	fieldSlackChanID  = "slack-channel-id"
	fieldSlackUGID    = "slack-usergroup-id"
	fieldUserID       = "user-id"
	fieldRotationID   = "rotation-id"
	fieldScheduleID   = "schedule-id"
)

type FieldValuePair App
type DestinationDisplayInfo App

func (a *App) FieldValuePair() graphql2.FieldValuePairResolver { return (*FieldValuePair)(a) }

func (a *FieldValuePair) Label(ctx context.Context, fvp *graphql2.FieldValuePair) (string, error) {
	if fvp.Label != "" {
		return fvp.Label, nil
	}

	app := (*App)(a)
	switch fvp.FieldID {
	case fieldSlackChanID:
		ch, err := app.SlackStore.Channel(ctx, fvp.Value)
		if err != nil {
			return "", err
		}
		return ch.Name, nil
	case fieldSlackUGID:
		ug, err := app.SlackStore.UserGroup(ctx, fvp.Value)
		if err != nil {
			return "", err
		}

		return ug.Handle, nil
	case fieldUserID:
		u, err := app.FindOneUser(ctx, fvp.Value)
		if err != nil {
			return "", err
		}
		return u.Name, nil
	case fieldRotationID:
		r, err := app.FindOneRotation(ctx, fvp.Value)
		if err != nil {
			return "", err
		}
		return r.Name, nil
	case fieldScheduleID:
		s, err := app.FindOneSchedule(ctx, fvp.Value)
		if err != nil {
			return "", err
		}
		return s.Name, nil
	}

	return "", validation.NewGenericError("unsupported fieldID")
}

func (q *Query) DestinationFieldValueName(ctx context.Context, input graphql2.DestinationFieldValidateInput) (string, error) {
	switch input.FieldID {
	case fieldSlackChanID:
		ch, err := q.SlackChannel(ctx, input.Value)
		if err != nil {
			return "", err
		}

		return ch.Name, nil
	case fieldSlackUGID:
		ug, err := q.SlackUserGroup(ctx, input.Value)
		if err != nil {
			return "", err
		}

		return ug.Handle, nil
	}

	return "", validation.NewGenericError("unsupported fieldID")
}

func (q *Query) DestinationFieldSearch(ctx context.Context, input graphql2.DestinationFieldSearchInput) (*graphql2.FieldValueConnection, error) {
	switch input.FieldID {
	case fieldSlackChanID:
		res, err := q.SlackChannels(ctx, &graphql2.SlackChannelSearchOptions{
			Omit:   input.Omit,
			First:  input.First,
			Search: input.Search,
			After:  input.After,
		})
		if err != nil {
			return nil, err
		}

		var nodes []graphql2.FieldValuePair
		for _, c := range res.Nodes {
			nodes = append(nodes, graphql2.FieldValuePair{
				FieldID: input.FieldID,
				Value:   c.ID,
				Label:   c.Name,
			})
		}

		return &graphql2.FieldValueConnection{
			Nodes:    nodes,
			PageInfo: res.PageInfo,
		}, nil
	case fieldSlackUGID:
		res, err := q.SlackUserGroups(ctx, &graphql2.SlackUserGroupSearchOptions{
			Omit:   input.Omit,
			First:  input.First,
			Search: input.Search,
			After:  input.After,
		})
		if err != nil {
			return nil, err
		}

		var nodes []graphql2.FieldValuePair
		for _, ug := range res.Nodes {
			nodes = append(nodes, graphql2.FieldValuePair{
				FieldID: input.FieldID,
				Value:   ug.ID,
				Label:   ug.Handle,
			})
		}

		return &graphql2.FieldValueConnection{
			Nodes:    nodes,
			PageInfo: res.PageInfo,
		}, nil
	}

	return nil, validation.NewGenericError("unsupported fieldID")
}

func (q *Query) DestinationFieldValidate(ctx context.Context, input graphql2.DestinationFieldValidateInput) (bool, error) {
	switch input.DestType {
	case destTwilioSMS, destTwilioVoice:
		if input.FieldID != fieldPhoneNumber {
			return false, validation.NewGenericError("unsupported field")
		}
		n, err := phonenumbers.Parse(input.Value, "")
		if err != nil {
			return false, nil
		}
		return phonenumbers.IsValidNumber(n), nil
	case destSMTP:
		if input.FieldID != fieldEmailAddress {
			return false, validation.NewGenericError("unsupported field")
		}

		return validate.Email("Email", input.Value) == nil, nil
	case destWebhook:
		if input.FieldID != fieldWebhookURL {
			return false, validation.NewGenericError("unsupported field")
		}

		err := validate.AbsoluteURL("URL", input.Value)
		return err == nil, nil
	}

	return false, validation.NewGenericError("unsupported data type")
}

func (q *Query) DestinationTypes(ctx context.Context) ([]graphql2.DestinationTypeInfo, error) {
	cfg := config.FromContext(ctx)
	types := []graphql2.DestinationTypeInfo{
		{
			Type:                  destTwilioSMS,
			Name:                  "Text Message (SMS)",
			Enabled:               cfg.Twilio.Enable,
			DisabledMessage:       "Twilio must be configured by an administrator",
			UserDisclaimer:        cfg.General.NotificationDisclaimer,
			SupportsStatusUpdates: true,
			IsContactMethod:       true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldPhoneNumber,
				Label:              "Phone Number",
				Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
				PlaceholderText:    "11235550123",
				Prefix:             "+",
				InputType:          "tel",
				SupportsValidation: true,
			}},
		},
		{
			Type:                  destTwilioVoice,
			Name:                  "Voice Call",
			Enabled:               cfg.Twilio.Enable,
			DisabledMessage:       "Twilio must be configured by an administrator",
			UserDisclaimer:        cfg.General.NotificationDisclaimer,
			IsContactMethod:       true,
			SupportsStatusUpdates: true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldPhoneNumber,
				Label:              "Phone Number",
				Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
				PlaceholderText:    "11235550123",
				Prefix:             "+",
				InputType:          "tel",
				SupportsValidation: true,
			}},
		},
		{
			Type:                  destSMTP,
			Name:                  "Email",
			Enabled:               cfg.SMTP.Enable,
			IsContactMethod:       true,
			SupportsStatusUpdates: true,
			DisabledMessage:       "SMTP must be configured by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldEmailAddress,
				Label:              "Email Address",
				PlaceholderText:    "foobar@example.com",
				InputType:          "email",
				SupportsValidation: true,
			}},
		},
		{
			Type:                  destWebhook,
			Name:                  "Webhook",
			Enabled:               cfg.Webhook.Enable,
			IsContactMethod:       true,
			IsEPTarget:            true,
			IsSchedOnCallNotify:   true,
			SupportsStatusUpdates: true,
			StatusUpdatesRequired: true,
			DisabledMessage:       "Webhooks must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldWebhookURL,
				Label:              "Webhook URL",
				PlaceholderText:    "https://example.com",
				InputType:          "url",
				Hint:               "Webhook Documentation",
				HintURL:            "/docs#webhooks",
				SupportsValidation: true,
			}},
		},
		{
			Type:                  destSlackDM,
			Name:                  "Slack Message (DM)",
			Enabled:               cfg.Slack.Enable,
			IsContactMethod:       true,
			SupportsStatusUpdates: true,
			StatusUpdatesRequired: true,
			DisabledMessage:       "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:         fieldSlackUserID,
				Label:           "Slack User",
				PlaceholderText: "member ID",
				InputType:       "text",
				// IsSearchSelectable: true, // TODO: implement search select functionality for users
				Hint: `Go to your Slack profile, click the three dots, and select "Copy member ID".`,
			}},
		},
		{
			Type:                  destSlackChan,
			Name:                  "Slack Channel",
			Enabled:               cfg.Slack.Enable,
			IsEPTarget:            true,
			IsSchedOnCallNotify:   true,
			SupportsStatusUpdates: true,
			StatusUpdatesRequired: true,
			DisabledMessage:       "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:        fieldSlackChanID,
				Label:          "Slack Channel",
				InputType:      "text",
				SupportsSearch: true,
			}},
		},
		{
			Type:                destSlackUG,
			Name:                "Update Slack User Group",
			Enabled:             cfg.Slack.Enable,
			IsSchedOnCallNotify: true,
			DisabledMessage:     "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:        fieldSlackUGID,
				Label:          "User Group",
				InputType:      "text",
				SupportsSearch: true,
				Hint:           "The selected group's membership will be replaced/set to the schedule's on-call user(s).",
			}, {
				FieldID:        fieldSlackChanID,
				Label:          "Slack Channel (for errors)",
				InputType:      "text",
				SupportsSearch: true,
				Hint:           "If the user group update fails, an error will be posted to this channel.",
			}},
		},
	}

	slices.SortStableFunc(types, func(a, b graphql2.DestinationTypeInfo) int {
		if a.Enabled && !b.Enabled {
			return -1
		}
		if !a.Enabled && b.Enabled {
			return 1
		}

		// keep order for types that are both enabled or both disabled
		return 0
	})

	return types, nil
}
