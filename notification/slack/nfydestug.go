package slack

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

var (
	_ nfydest.Provider      = (*UserGroupSender)(nil)
	_ nfydest.FieldSearcher = (*UserGroupSender)(nil)
)

func (ug *UserGroupSender) ID() string { return DestTypeSlackUsergroup }
func (ug *UserGroupSender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                 DestTypeSlackUsergroup,
		Name:                 "Update Slack User Group",
		Enabled:              cfg.Slack.Enable,
		SupportsOnCallNotify: true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:        FieldSlackUsergroupID,
			Label:          "User Group",
			InputType:      "text",
			SupportsSearch: true,
			Hint:           "The selected group's membership will be replaced/set to the schedule's on-call user(s).",
		}, {
			FieldID:        FieldSlackChannelID,
			Label:          "Slack Channel (for errors)",
			InputType:      "text",
			SupportsSearch: true,
			Hint:           "If the user group update fails, an error will be posted to this channel.",
		}},
	}, nil
}

func (ug *UserGroupSender) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldSlackUsergroupID:
		return ug.ValidateUserGroup(ctx, value)
	case FieldSlackChannelID:
		return ug.ValidateChannel(ctx, value)
	}

	return validation.NewGenericError("unknown field ID")
}

func (ug *UserGroupSender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	u, err := ug.UserGroup(ctx, args[FieldSlackUsergroupID])
	if err != nil {
		return nil, err
	}

	team, err := ug.Team(ctx, u.TeamID)
	if err != nil {
		return nil, err
	}

	if team.IconURL == "" {
		team.IconURL = "builtin://slack"
	}

	return &nfydest.DisplayInfo{
		IconURL:     team.IconURL,
		IconAltText: team.Name,
		Text:        u.Handle,
	}, nil
}

func (ug *UserGroupSender) SearchField(ctx context.Context, fieldID string, options nfydest.SearchOptions) (*nfydest.SearchResult, error) {
	switch fieldID {
	case FieldSlackChannelID:
		return ug.ChannelSender.SearchField(ctx, fieldID, options)
	case FieldSlackUsergroupID:
		return nfydest.SearchByListFunc(ctx, options, ug.ListUserGroups)
	}

	return nil, validation.NewGenericError("unsupported field ID")
}

func (ug *UserGroupSender) FieldLabel(ctx context.Context, fieldID, value string) (string, error) {
	switch fieldID {
	case FieldSlackChannelID:
		return ug.ChannelSender.FieldLabel(ctx, fieldID, value)
	case FieldSlackUsergroupID:
		grp, err := ug.UserGroup(ctx, value)
		if err != nil {
			return "", err
		}

		return grp.Handle, nil
	}

	return "", validation.NewGenericError("unsupported field ID")
}
