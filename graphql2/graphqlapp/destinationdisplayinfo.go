package graphqlapp

import (
	"context"
	"net/mail"
	"net/url"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
)

type (
	Destination App
)

func (a *App) Destination() graphql2.DestinationResolver { return (*Destination)(a) }

// DisplayInfo will return the display information for a destination by mapping to Query.DestinationDisplayInfo.
func (a *Destination) DisplayInfo(ctx context.Context, obj *graphql2.Destination) (*graphql2.DestinationDisplayInfo, error) {
	if obj.DisplayInfo != nil {
		return obj.DisplayInfo, nil
	}

	values := make([]graphql2.FieldValueInput, len(obj.Values))
	for i, v := range obj.Values {
		values[i] = graphql2.FieldValueInput{FieldID: v.FieldID, Value: v.Value}
	}

	return (*Query)(a).DestinationDisplayInfo(ctx, graphql2.DestinationInput{Type: obj.Type, Values: values})
}

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
			LinkURL:     team.UserLink(u.ID),
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
			LinkURL:     team.ChannelLink(ch.ID),
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
			Text:        ug.Handle,
		}, nil
	}

	return nil, validation.NewGenericError("unsupported data type")
}
