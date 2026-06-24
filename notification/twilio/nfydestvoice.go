package twilio

import (
	"context"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

const (
	DestTypeTwilioVoice  = "builtin-twilio-voice"
	FallbackIconURLVoice = "builtin://phone-voice"
)

var _ nfydest.Provider = (*Voice)(nil)

func (v *Voice) ID() string { return DestTypeTwilioVoice }
func (v *Voice) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeTwilioVoice,
		Name:                       "Voice Call",
		Enabled:                    cfg.Twilio.Enable,
		UserDisclaimer:             cfg.General.NotificationDisclaimer,
		SupportsAlertNotifications: true,
		SupportsUserVerification:   true,
		SupportsStatusUpdates:      true,
		UserVerificationRequired:   true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldPhoneNumber,
			Label:              "Phone Number",
			Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
			PlaceholderText:    "11235550123",
			Prefix:             "+",
			InputType:          "tel",
			SupportsValidation: true,
		}},
	}, nil
}

func (v *Voice) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldPhoneNumber:
		n, err := phonenumbers.Parse(value, "")
		if err != nil {
			return validation.WrapError(err)
		}
		if !phonenumbers.IsValidNumber(n) {
			return validation.NewGenericError("invalid phone number")
		}
		return nil
	}

	return validation.NewGenericError("unknown field ID")
}

func (v *Voice) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	n, err := phonenumbers.Parse(args[FieldPhoneNumber], "")
	if err != nil {
		return nil, validation.WrapError(err)
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURLVoice,
		IconAltText: "Voice Call",
		Text:        phonenumbers.Format(n, phonenumbers.INTERNATIONAL),
	}, nil
}
