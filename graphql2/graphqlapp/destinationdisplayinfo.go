package graphqlapp

import (
	"context"
	"net/mail"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/util/errutil"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation"
)

type (
	Destination      App
	DestinationInput App
)

func (a *App) Destination() graphql2.DestinationResolver           { return (*Destination)(a) }
func (a *App) DestinationInput() graphql2.DestinationInputResolver { return (*DestinationInput)(a) }

func (a *DestinationInput) Values(ctx context.Context, obj *gadb.DestV1, values []graphql2.FieldValueInput) error {
	obj.Args = make(map[string]string, len(values))
	for _, val := range values {
		obj.Args[val.FieldID] = val.Value
	}
	return nil
}

func (a *Destination) Values(ctx context.Context, obj *gadb.DestV1) ([]graphql2.FieldValuePair, error) {
	if obj.Args != nil {
		pairs := make([]graphql2.FieldValuePair, 0, len(obj.Args))
		for k, v := range obj.Args {
			pairs = append(pairs, graphql2.FieldValuePair{FieldID: k, Value: v})
		}
		return pairs, nil
	}

	return nil, nil
}

// DisplayInfo will return the display information for a destination by mapping to Query.DestinationDisplayInfo.
func (a *Destination) DisplayInfo(ctx context.Context, obj *gadb.DestV1) (graphql2.InlineDisplayInfo, error) {
	info, err := (*Query)(a)._DestinationDisplayInfo(ctx, gadb.DestV1{Type: obj.Type, Args: obj.Args}, true)
	if err != nil {
		isUnsafe, safeErr := errutil.ScrubError(err)
		if isUnsafe {
			log.Log(ctx, err)
		}
		return &graphql2.DestinationDisplayInfoError{Error: safeErr.Error()}, nil
	}

	return info, nil
}

func (a *Query) DestinationDisplayInfo(ctx context.Context, dest gadb.DestV1) (*nfydest.DisplayInfo, error) {
	return a._DestinationDisplayInfo(ctx, dest, false)
}

func (a *Query) _DestinationDisplayInfo(ctx context.Context, dest gadb.DestV1, skipValidation bool) (*nfydest.DisplayInfo, error) {
	app := (*App)(a)
	cfg := config.FromContext(ctx)
	if !skipValidation {
		if err := app.ValidateDestination(ctx, "input", &dest); err != nil {
			return nil, err
		}
	}
	switch dest.Type {
	case destTwilioSMS:
		n, err := phonenumbers.Parse(dest.Arg(fieldPhoneNumber), "")
		if err != nil {
			return nil, validation.WrapError(err)
		}

		return &nfydest.DisplayInfo{
			IconURL:     "builtin://phone-text",
			IconAltText: "Text Message",
			Text:        phonenumbers.Format(n, phonenumbers.INTERNATIONAL),
		}, nil
	case destTwilioVoice:
		n, err := phonenumbers.Parse(dest.Arg(fieldPhoneNumber), "")
		if err != nil {
			return nil, validation.WrapError(err)
		}
		return &nfydest.DisplayInfo{
			IconURL:     "builtin://phone-voice",
			IconAltText: "Voice Call",
			Text:        phonenumbers.Format(n, phonenumbers.INTERNATIONAL),
		}, nil
	case destSMTP:
		e, err := mail.ParseAddress(dest.Arg(fieldEmailAddress))
		if err != nil {
			return nil, validation.WrapError(err)
		}
		return &nfydest.DisplayInfo{
			IconURL:     "builtin://email",
			IconAltText: "Email",
			Text:        e.Address,
		}, nil
	case destRotation:
		r, err := app.FindOneRotation(ctx, dest.Arg(fieldRotationID))
		if err != nil {
			return nil, err
		}
		return &nfydest.DisplayInfo{
			IconURL:     "builtin://rotation",
			IconAltText: "Rotation",
			LinkURL:     cfg.CallbackURL("/rotations/" + r.ID),
			Text:        r.Name,
		}, nil
	case destSchedule:
		s, err := app.FindOneSchedule(ctx, dest.Arg(fieldScheduleID))
		if err != nil {
			return nil, err
		}
		return &nfydest.DisplayInfo{
			IconURL:     "builtin://schedule",
			IconAltText: "Schedule",
			LinkURL:     cfg.CallbackURL("/schedules/" + s.ID),
			Text:        s.Name,
		}, nil
	case destUser:
		u, err := app.FindOneUser(ctx, dest.Arg(fieldUserID))
		if err != nil {
			return nil, err
		}
		return &nfydest.DisplayInfo{
			IconURL:     cfg.CallbackURL("/api/v2/user-avatar/" + u.ID),
			IconAltText: "User",
			LinkURL:     cfg.CallbackURL("/users/" + u.ID),
			Text:        u.Name,
		}, nil
	}

	return app.DestReg.DisplayInfo(ctx, dest)
}
