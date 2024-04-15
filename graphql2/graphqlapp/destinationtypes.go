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
	case fieldRotationID:
		rot, err := q.Rotation(ctx, input.Value)
		if err != nil {
			return "", err
		}

		return rot.Name, nil
	case fieldScheduleID:
		sched, err := q.Schedule(ctx, input.Value)
		if err != nil {
			return "", err
		}

		return sched.Name, nil
	case fieldUserID:
		u, err := q.User(ctx, &input.Value)
		if err != nil {
			return "", err
		}
		return u.Name, nil
	}

	return "", validation.NewGenericError("unsupported fieldID")
}

func (q *Query) DestinationFieldSearch(ctx context.Context, input graphql2.DestinationFieldSearchInput) (*graphql2.FieldSearchConnection, error) {
	favFirst := true

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

		var nodes []graphql2.FieldSearchResult
		for _, c := range res.Nodes {
			nodes = append(nodes, graphql2.FieldSearchResult{
				FieldID: input.FieldID,
				Value:   c.ID,
				Label:   c.Name,
			})
		}

		return &graphql2.FieldSearchConnection{
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

		var nodes []graphql2.FieldSearchResult
		for _, ug := range res.Nodes {
			nodes = append(nodes, graphql2.FieldSearchResult{
				FieldID: input.FieldID,
				Value:   ug.ID,
				Label:   ug.Handle,
			})
		}

		return &graphql2.FieldSearchConnection{
			Nodes:    nodes,
			PageInfo: res.PageInfo,
		}, nil
	case fieldRotationID:
		res, err := q.Rotations(ctx, &graphql2.RotationSearchOptions{
			Omit:           input.Omit,
			First:          input.First,
			Search:         input.Search,
			After:          input.After,
			FavoritesFirst: &favFirst,
		})
		if err != nil {
			return nil, err
		}

		var nodes []graphql2.FieldSearchResult
		for _, rot := range res.Nodes {
			nodes = append(nodes, graphql2.FieldSearchResult{
				FieldID:    input.FieldID,
				Value:      rot.ID,
				Label:      rot.Name,
				IsFavorite: rot.IsUserFavorite(),
			})
		}

		return &graphql2.FieldSearchConnection{
			Nodes:    nodes,
			PageInfo: res.PageInfo,
		}, nil
	case fieldScheduleID:
		res, err := q.Schedules(ctx, &graphql2.ScheduleSearchOptions{
			Omit:           input.Omit,
			First:          input.First,
			Search:         input.Search,
			After:          input.After,
			FavoritesFirst: &favFirst,
		})
		if err != nil {
			return nil, err
		}

		var nodes []graphql2.FieldSearchResult
		for _, sched := range res.Nodes {
			nodes = append(nodes, graphql2.FieldSearchResult{
				FieldID:    input.FieldID,
				Value:      sched.ID,
				Label:      sched.Name,
				IsFavorite: sched.IsUserFavorite(),
			})
		}

		return &graphql2.FieldSearchConnection{
			Nodes:    nodes,
			PageInfo: res.PageInfo,
		}, nil
	case fieldUserID:
		res, err := q.Users(ctx, &graphql2.UserSearchOptions{
			Omit:           input.Omit,
			First:          input.First,
			Search:         input.Search,
			After:          input.After,
			FavoritesFirst: &favFirst,
		}, input.First, input.After, input.Search)
		if err != nil {
			return nil, err
		}

		var nodes []graphql2.FieldSearchResult
		for _, u := range res.Nodes {
			nodes = append(nodes, graphql2.FieldSearchResult{
				FieldID:    input.FieldID,
				Value:      u.ID,
				Label:      u.Name,
				IsFavorite: u.IsUserFavorite(),
			})
		}

		return &graphql2.FieldSearchConnection{
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
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:         fieldSlackUserID,
				Label:           "Slack User",
				PlaceholderText: "member ID",
				InputType:       "text",
				// supportsSearch: true, // TODO: implement search select functionality for users
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
		{
			Type:       destRotation,
			Name:       "Rotation",
			Enabled:    true,
			IsEPTarget: true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:        fieldRotationID,
				Label:          "Rotation",
				InputType:      "text",
				SupportsSearch: true,
			}},
		},
		{
			Type:       destSchedule,
			Name:       "Schedule",
			Enabled:    true,
			IsEPTarget: true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:        fieldScheduleID,
				Label:          "Schedule",
				InputType:      "text",
				SupportsSearch: true,
			}},
		},
		{
			Type:       destUser,
			Name:       "User",
			Enabled:    true,
			IsEPTarget: true,
			RequiredFields: []graphql2.DestinationFieldConfig{{
				FieldID:        fieldUserID,
				Label:          "User",
				InputType:      "text",
				SupportsSearch: true,
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
