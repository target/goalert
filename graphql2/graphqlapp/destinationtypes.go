package graphqlapp

import (
	"context"
	"net/mail"
	"net/url"
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

func (a *Query) DestinationDisplayInfo(ctx context.Context, dest graphql2.DestinationInput) (*graphql2.DestinationDisplayInfo, error) {
	app := (*App)(a)
	cfg := config.FromContext(ctx)
	switch dest.Type {
	case destTwilioSMS:
		n, err := phonenumbers.Parse(dest.FieldValue(fieldPhoneNumber), "")
		if err != nil {
			return nil, validation.WrapError(err)
		}

		return &graphql2.DestinationDisplayInfo{
			IconURL:     "builtin://phone-text",
			IconAltText: "Text Message",
			Text:        phonenumbers.Format(n, phonenumbers.INTERNATIONAL),
		}, nil
	case destTwilioVoice:
		n, err := phonenumbers.Parse(dest.FieldValue(fieldPhoneNumber), "")
		if err != nil {
			return nil, validation.WrapError(err)
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     "builtin://phone-voice",
			IconAltText: "Voice Call",
			Text:        phonenumbers.Format(n, phonenumbers.INTERNATIONAL),
		}, nil
	case destSMTP:
		e, err := mail.ParseAddress(dest.FieldValue(fieldEmailAddress))
		if err != nil {
			return nil, validation.WrapError(err)
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     "builtin://email",
			IconAltText: "Email",
			Text:        e.Address,
		}, nil
	case destRotation:
		r, err := app.FindOneRotation(ctx, dest.FieldValue(fieldRotationID))
		if err != nil {
			return nil, err
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     "builtin://rotation",
			IconAltText: "Rotation",
			LinkURL:     cfg.CallbackURL("/rotations/" + r.ID),
			Text:        r.Name,
		}, nil
	case destSchedule:
		s, err := app.FindOneSchedule(ctx, dest.FieldValue(fieldScheduleID))
		if err != nil {
			return nil, err
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     "builtin://schedule",
			IconAltText: "Schedule",
			LinkURL:     cfg.CallbackURL("/schedules/" + s.ID),
			Text:        s.Name,
		}, nil
	case destUser:
		u, err := app.FindOneUser(ctx, dest.FieldValue(fieldUserID))
		if err != nil {
			return nil, err
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     cfg.CallbackURL("/api/v2/user-avatar/" + u.ID),
			IconAltText: "User",
			LinkURL:     cfg.CallbackURL("/users/" + u.ID),
			Text:        u.Name,
		}, nil

	case destWebhook:
		u, err := url.Parse(dest.FieldValue(fieldWebhookURL))
		if err != nil {
			return nil, validation.WrapError(err)
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     "builtin://webhook",
			IconAltText: "Webhook",
			Text:        u.Hostname(),
		}, nil
	case destSlackDM:
		u, err := app.SlackStore.User(ctx, dest.FieldValue(fieldSlackUserID))
		if err != nil {
			return nil, err
		}

		team, err := app.SlackStore.Team(ctx, u.TeamID)
		if err != nil {
			return nil, err
		}

		if team.IconURL == "" {
			team.IconURL = "builtin://slack"
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     team.IconURL,
			IconAltText: team.Name,
			LinkURL:     "https://app.slack.com/client/" + url.PathEscape(team.ID) + "/" + url.PathEscape(u.ID),
			Text:        u.Name,
		}, nil
	case destSlackChan:
		ch, err := app.SlackStore.Channel(ctx, dest.FieldValue(fieldSlackChanID))
		if err != nil {
			return nil, err
		}

		team, err := app.SlackStore.Team(ctx, ch.TeamID)
		if err != nil {
			return nil, err
		}

		if team.IconURL == "" {
			team.IconURL = "builtin://slack"
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     team.IconURL,
			IconAltText: team.Name,
			LinkURL:     "https://app.slack.com/client/" + url.PathEscape(team.ID) + "/" + url.PathEscape(ch.ID),
			Text:        ch.Name,
		}, nil

	case destSlackUG:
		ug, err := app.SlackStore.UserGroup(ctx, dest.FieldValue(fieldSlackUGID))
		if err != nil {
			return nil, err
		}

		team, err := app.SlackStore.Team(ctx, ug.TeamID)
		if err != nil {
			return nil, err
		}

		if team.IconURL == "" {
			team.IconURL = "builtin://slack"
		}
		return &graphql2.DestinationDisplayInfo{
			IconURL:     team.IconURL,
			IconAltText: team.Name,
			LinkURL:     "https://app.slack.com/client/" + url.PathEscape(team.ID) + "/manage/usergroups",
			Text:        ug.Handle,
		}, nil
	}

	return nil, validation.NewGenericError("unsupported data type")
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
			Type:            destTwilioSMS,
			Name:            "Text Message (SMS)",
			Enabled:         cfg.Twilio.Enable,
			DisabledMessage: "Twilio must be configured by an administrator",
			UserDisclaimer:  cfg.General.NotificationDisclaimer,
			IsContactMethod: true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldPhoneNumber,
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
			Type:            destTwilioVoice,
			Name:            "Voice Call",
			Enabled:         cfg.Twilio.Enable,
			DisabledMessage: "Twilio must be configured by an administrator",
			UserDisclaimer:  cfg.General.NotificationDisclaimer,
			IsContactMethod: true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldPhoneNumber,
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
			Type:            destSMTP,
			Name:            "Email",
			Enabled:         cfg.SMTP.Enable,
			IsContactMethod: true,
			DisabledMessage: "SMTP must be configured by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldEmailAddress,
				LabelSingular:      "Email Address",
				LabelPlural:        "Email Addresses",
				PlaceholderText:    "foobar@example.com",
				InputType:          "email",
				SupportsValidation: true,
			}},
		},
		{
			Type:                destWebhook,
			Name:                "Webhook",
			Enabled:             cfg.Webhook.Enable,
			IsContactMethod:     true,
			IsEPTarget:          true,
			IsSchedOnCallNotify: true,
			DisabledMessage:     "Webhooks must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldWebhookURL,
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
			Type:            destSlackDM,
			Name:            "Slack Message (DM)",
			Enabled:         cfg.Slack.Enable,
			IsContactMethod: true,
			DisabledMessage: "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:         fieldSlackUserID,
				LabelSingular:   "Slack User",
				LabelPlural:     "Slack Users",
				PlaceholderText: "member ID",
				InputType:       "text",
				// IsSearchSelectable: true, // TODO: implement search select functionality for users
				Hint: `Go to your Slack profile, click the three dots, and select "Copy member ID".`,
			}},
		},
		{
			Type:                destSlackChan,
			Name:                "Slack Channel",
			Enabled:             cfg.Slack.Enable,
			IsEPTarget:          true,
			IsSchedOnCallNotify: true,
			DisabledMessage:     "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldSlackChanID,
				LabelSingular:      "Slack Channel",
				LabelPlural:        "Slack Channels",
				InputType:          "text",
				IsSearchSelectable: true,
			}},
		},
		{
			Type:                destSlackUG,
			Name:                "Update Slack User Group",
			Enabled:             cfg.Slack.Enable,
			IsSchedOnCallNotify: true,
			DisabledMessage:     "Slack must be enabled by an administrator",
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:            fieldSlackUGID,
				LabelSingular:      "User Group",
				LabelPlural:        "User Groups",
				InputType:          "text",
				IsSearchSelectable: true,
				Hint:               "The selected group's membership will be replaced/set to the schedule's on-call user(s).",
			}, {
				FieldID:            fieldSlackChanID,
				LabelSingular:      "Slack Channel (for errors)",
				LabelPlural:        "Slack Channels (for errors)",
				InputType:          "text",
				IsSearchSelectable: true,
				Hint:               "If the user group update fails, an error will be posted to this channel.",
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
