package email

import (
	"context"
	"net/mail"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

const (
	DestTypeEmail     = "builtin-smtp-email"
	FieldEmailAddress = "email_address"
	FallbackIconURL   = "builtin://email"
)

var _ nfydest.Provider = (*Sender)(nil)

func NewEmailDest(address string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeEmail, FieldEmailAddress, address)
}

func (s *Sender) ID() string { return DestTypeEmail }
func (s *Sender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeEmail,
		Name:                       "Email",
		Enabled:                    cfg.SMTP.Enable,
		SupportsAlertNotifications: true,
		SupportsUserVerification:   true,
		SupportsStatusUpdates:      true,
		UserVerificationRequired:   true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldEmailAddress,
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
	}, nil
}

func (s *Sender) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldEmailAddress:
		_, err := mail.ParseAddress(value)
		if err != nil {
			return validation.WrapError(err)
		}

		return nil
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *Sender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	e, err := mail.ParseAddress(args[FieldEmailAddress])
	if err != nil {
		return nil, validation.WrapError(err)
	}
	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Email",
		Text:        e.Address,
	}, nil
}
