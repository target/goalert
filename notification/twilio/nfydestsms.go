package twilio

import (
	"context"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

const (
	DestTypeTwilioSMS  = "builtin-twilio-sms"
	FieldPhoneNumber   = "phone_number"
	FallbackIconURLSMS = "builtin://phone-text"
)

var _ nfydest.Provider = (*SMS)(nil)

func (s *SMS) ID() string { return DestTypeTwilioSMS }
func (s *SMS) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeTwilioSMS,
		Name:                       "Text Message (SMS)",
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

func (s *SMS) ValidateField(ctx context.Context, fieldID, value string) error {
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

func (s *SMS) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	n, err := phonenumbers.Parse(args[FieldPhoneNumber], "")
	if err != nil {
		return nil, validation.WrapError(err)
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURLSMS,
		IconAltText: "Text Message",
		Text:        phonenumbers.Format(n, phonenumbers.INTERNATIONAL),
	}, nil
}
