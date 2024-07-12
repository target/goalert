package slack

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

var _ nfydest.Provider = (*DMSender)(nil)

func (dm *DMSender) ID() string { return DestTypeSlackDirectMessage }
func (dm *DMSender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeSlackDirectMessage,
		Name:                       "Slack Message (DM)",
		Enabled:                    cfg.Slack.Enable,
		SupportsAlertNotifications: true,
		SupportsUserVerification:   true,
		SupportsStatusUpdates:      true,
		UserVerificationRequired:   true,
		StatusUpdatesRequired:      true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:         FieldSlackUserID,
			Label:           "Slack User",
			PlaceholderText: "member ID",
			InputType:       "text",
			// supportsSearch: true, // TODO: implement search select functionality for users
			Hint: `Go to your Slack profile, click the three dots, and select "Copy member ID".`,
		}},
	}, nil
}

func (dm *DMSender) ValidateField(ctx context.Context, fieldID, value string) (bool, error) {
	switch fieldID {
	case FieldSlackChannelID:
		err := dm.ValidateUser(ctx, value)
		if validation.IsValidationError(err) {
			return false, nil
		}

		return err == nil, err
	}

	return false, validation.NewGenericError("unknown field ID")
}

func (dm *DMSender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	u, err := dm.User(ctx, args[FieldSlackUserID])
	if err != nil {
		return nil, err
	}

	team, err := dm.Team(ctx, u.TeamID)
	if err != nil {
		return nil, err
	}

	if team.IconURL == "" {
		team.IconURL = "builtin://slack"
	}

	return &nfydest.DisplayInfo{
		IconURL:     team.IconURL,
		IconAltText: team.Name,
		LinkURL:     team.UserLink(u.ID),
		Text:        u.Name,
	}, nil
}
