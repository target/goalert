package slack

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

var (
	_ nfydest.Provider      = (*ChannelSender)(nil)
	_ nfydest.FieldSearcher = (*ChannelSender)(nil)
)

func (s *ChannelSender) ID() string { return DestTypeSlackChannel }
func (s *ChannelSender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeSlackChannel,
		Name:                       "Slack Channel",
		Enabled:                    cfg.Slack.Enable,
		SupportsAlertNotifications: true,
		SupportsStatusUpdates:      true,
		SupportsOnCallNotify:       true,
		StatusUpdatesRequired:      true,
		SupportsSignals:            true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:        FieldSlackChannelID,
			Label:          "Slack Channel",
			InputType:      "text",
			SupportsSearch: true,
		}},
		DynamicParams: []nfydest.DynamicParamConfig{{
			ParamID: "message",
			Label:   "Message",
			Hint:    "The text of the message to send.",
		}},
	}, nil
}

func (s *ChannelSender) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldSlackChannelID:
		return s.ValidateChannel(ctx, value)
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *ChannelSender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	ch, err := s.Channel(ctx, args[FieldSlackChannelID])
	if err != nil {
		return nil, err
	}

	team, err := s.Team(ctx, ch.TeamID)
	if err != nil {
		return nil, err
	}

	if team.IconURL == "" {
		team.IconURL = FallbackIconURL
	}
	return &nfydest.DisplayInfo{
		IconURL:     team.IconURL,
		IconAltText: team.Name,
		LinkURL:     team.ChannelLink(ch.ID),
		Text:        ch.Name,
	}, nil
}

func (s *ChannelSender) SearchField(ctx context.Context, fieldID string, options nfydest.SearchOptions) (*nfydest.SearchResult, error) {
	switch fieldID {
	case FieldSlackChannelID:
		ch, err := s.ListChannels(ctx)
		if err != nil {
			return nil, err
		}
		return nfydest.SearchByList(ch, options, func(ch Channel) nfydest.FieldValue {
			return nfydest.FieldValue{
				Label: ch.Name,
				Value: ch.ID,
			}
		})
	}

	return nil, validation.NewGenericError("unsupported field ID")
}

func (s *ChannelSender) FieldLabel(ctx context.Context, fieldID, value string) (string, error) {
	switch fieldID {
	case FieldSlackChannelID:
		ch, err := s.Channel(ctx, value)
		if err != nil {
			return "", err
		}

		return ch.Name, nil
	}

	return "", validation.NewGenericError("unsupported field ID")
}
