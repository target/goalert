package graphqlapp

import (
	"context"
	"slices"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/nfydest"
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
	destAlert       = "builtin-alert"

	fieldPhoneNumber  = "phone_number"
	fieldEmailAddress = "email_address"
	fieldWebhookURL   = "webhook_url"
	fieldSlackUserID  = "slack_user_id"
	fieldSlackChanID  = "slack_channel_id"
	fieldSlackUGID    = "slack_usergroup_id"
	fieldUserID       = "user_id"
	fieldRotationID   = "rotation_id"
	fieldScheduleID   = "schedule_id"
)

type (
	FieldValuePair         App
	DestinationDisplayInfo App
)

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

	return q.DestReg.FieldLabel(ctx, input.DestType, input.FieldID, input.Value)
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

	var opts nfydest.SearchOptions
	opts.Omit = input.Omit
	if input.First != nil {
		opts.Limit = *input.First
	}
	if input.After != nil {
		opts.Cursor = *input.After
	}
	if input.Search != nil {
		opts.Search = *input.Search
	}

	res, err := q.DestReg.SearchField(ctx, input.DestType, input.FieldID, opts)
	if err != nil {
		return nil, err
	}
	var nodes []graphql2.FieldSearchResult
	for _, v := range res.Values {
		nodes = append(nodes, graphql2.FieldSearchResult{
			FieldID:    input.FieldID,
			Value:      v.Value,
			Label:      v.Label,
			IsFavorite: v.IsFavorite,
		})
	}

	return &graphql2.FieldSearchConnection{
		Nodes: nodes,
		PageInfo: &graphql2.PageInfo{
			HasNextPage: res.HasNextPage,
			EndCursor:   &res.Cursor,
		},
	}, nil
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

	return q.DestReg.ValidateField(ctx, input.DestType, input.FieldID, input.Value)
}

func (q *Query) DestinationTypes(ctx context.Context, isDynamicAction *bool) ([]nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	types := []nfydest.TypeInfo{
		{
			Type:            destAlert,
			Name:            "Alert",
			Enabled:         true,
			SupportsSignals: true,
			DynamicParams: []nfydest.DynamicParamConfig{{
				ParamID: "summary",
				Label:   "Summary",
				Hint:    "Short summary of the alert (used for things like SMS).",
			}, {
				ParamID: "details",
				Label:   "Details",
				Hint:    "Full body (markdown) text of the alert.",
			}, {
				ParamID: "dedup",
				Label:   "Dedup",
				Hint:    "Stable identifier for de-duplication and closing existing alerts.",
			}, {
				ParamID: "close",
				Label:   "Close",
				Hint:    "If true, close an existing alert.",
			}},
		},
		{
			Type:                       destTwilioSMS,
			Name:                       "Text Message (SMS)",
			Enabled:                    cfg.Twilio.Enable,
			UserDisclaimer:             cfg.General.NotificationDisclaimer,
			SupportsAlertNotifications: true,
			SupportsUserVerification:   true,
			SupportsStatusUpdates:      true,
			UserVerificationRequired:   true,
			RequiredFields: []nfydest.FieldConfig{{
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
			Type:                       destTwilioVoice,
			Name:                       "Voice Call",
			Enabled:                    cfg.Twilio.Enable,
			UserDisclaimer:             cfg.General.NotificationDisclaimer,
			SupportsAlertNotifications: true,
			SupportsUserVerification:   true,
			SupportsStatusUpdates:      true,
			UserVerificationRequired:   true,
			RequiredFields: []nfydest.FieldConfig{{
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
			Type:                       destSMTP,
			Name:                       "Email",
			Enabled:                    cfg.SMTP.Enable,
			SupportsAlertNotifications: true,
			SupportsUserVerification:   true,
			SupportsStatusUpdates:      true,
			UserVerificationRequired:   true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:            fieldEmailAddress,
				Label:              "Email Address",
				PlaceholderText:    "foobar@example.com",
				InputType:          "email",
				SupportsValidation: true,
			}},
			DynamicParams: []nfydest.DynamicParamConfig{{
				ParamID: "subject",
				Label:   "Subject",
				Hint:    "Subject of the email message.",
			}, {
				ParamID: "body",
				Label:   "Body",
				Hint:    "Body of the email message.",
			}},
		},
		{
			Type:                       destWebhook,
			Name:                       "Webhook",
			Enabled:                    cfg.Webhook.Enable,
			SupportsUserVerification:   true,
			SupportsOnCallNotify:       true,
			SupportsSignals:            true,
			SupportsStatusUpdates:      true,
			SupportsAlertNotifications: true,
			StatusUpdatesRequired:      true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:            fieldWebhookURL,
				Label:              "Webhook URL",
				PlaceholderText:    "https://example.com",
				InputType:          "url",
				Hint:               "Webhook Documentation",
				HintURL:            "/docs#webhooks",
				SupportsValidation: true,
			}},
			DynamicParams: []nfydest.DynamicParamConfig{
				{
					ParamID: "body",
					Label:   "Body",
					Hint:    "The body of the request.",
				},
				{
					ParamID:      "content-type",
					Label:        "Content Type",
					Hint:         "The content type (e.g., application/json).",
					DefaultValue: `"application/json"`, // Because this is an expression, it needs the double quotes.
				},
			},
		},
		{
			Type:                       destSlackDM,
			Name:                       "Slack Message (DM)",
			Enabled:                    cfg.Slack.Enable,
			SupportsAlertNotifications: true,
			SupportsUserVerification:   true,
			SupportsStatusUpdates:      true,
			UserVerificationRequired:   true,
			StatusUpdatesRequired:      true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:         fieldSlackUserID,
				Label:           "Slack User",
				PlaceholderText: "member ID",
				InputType:       "text",
				// supportsSearch: true, // TODO: implement search select functionality for users
				Hint: `Go to your Slack profile, click the three dots, and select "Copy member ID".`,
			}},
		},
		{
			Type:                       destSlackChan,
			Name:                       "Slack Channel",
			Enabled:                    cfg.Slack.Enable,
			SupportsAlertNotifications: true,
			SupportsStatusUpdates:      true,
			SupportsOnCallNotify:       true,
			StatusUpdatesRequired:      true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:        fieldSlackChanID,
				Label:          "Slack Channel",
				InputType:      "text",
				SupportsSearch: true,
			}},
			DynamicParams: []nfydest.DynamicParamConfig{{
				ParamID: "message",
				Label:   "Message",
				Hint:    "The text of the message to send.",
			}},
		},
		{
			Type:                 destSlackUG,
			Name:                 "Update Slack User Group",
			Enabled:              cfg.Slack.Enable,
			SupportsOnCallNotify: true,
			RequiredFields: []nfydest.FieldConfig{{
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
			Type:                       destRotation,
			Name:                       "Rotation",
			Enabled:                    true,
			SupportsAlertNotifications: true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:        fieldRotationID,
				Label:          "Rotation",
				InputType:      "text",
				SupportsSearch: true,
			}},
		},
		{
			Type:                       destSchedule,
			Name:                       "Schedule",
			Enabled:                    true,
			SupportsAlertNotifications: true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:        fieldScheduleID,
				Label:          "Schedule",
				InputType:      "text",
				SupportsSearch: true,
			}},
		},
		{
			Type:                       destUser,
			Name:                       "User",
			Enabled:                    true,
			SupportsAlertNotifications: true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:        fieldUserID,
				Label:          "User",
				InputType:      "text",
				SupportsSearch: true,
			}},
		},
	}

	fromReg, err := q.DestReg.Types(ctx)
	if err != nil {
		return nil, err
	}
	types = append(types, fromReg...)

	slices.SortStableFunc(types, func(a, b nfydest.TypeInfo) int {
		if a.Enabled && !b.Enabled {
			return -1
		}
		if !a.Enabled && b.Enabled {
			return 1
		}

		// keep order for types that are both enabled or both disabled
		return 0
	})

	filtered := types[:0]
	for _, t := range types {
		if isDynamicAction != nil && *isDynamicAction != t.IsDynamicAction() {
			continue
		}

		filtered = append(filtered, t)
	}

	return filtered, nil
}
