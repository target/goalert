package webhook

import (
	"context"
	"mime"
	"net/url"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const (
	DestTypeWebhook       = "builtin-webhook"
	DestTypeCustomWebhook = "builtin-custom-webhook"
	FieldWebhookURL       = "webhook_url"
	FieldBodyTemplate     = "body_template"
	FieldContentType      = "content_type"
	ParamBody             = "body"
	ParamContentType      = "content_type"
	FallbackIconURL       = "builtin://webhook"
)

func NewWebhookDest(url string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeWebhook, FieldWebhookURL, url)
}

func NewCustomWebhookDest(url, bodyTemplate, contentType string) gadb.DestV1 {
	return gadb.NewDestV1(
		DestTypeCustomWebhook,
		FieldWebhookURL, url,
		FieldBodyTemplate, bodyTemplate,
		FieldContentType, contentType,
	)
}

var _ (nfydest.Provider) = (*Sender)(nil)
var _ (nfydest.Provider) = (*CustomSender)(nil)

func (Sender) ID() string { return DestTypeWebhook }

func (CustomSender) ID() string { return DestTypeCustomWebhook }

func (Sender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeWebhook,
		Name:                       "Webhook",
		Enabled:                    cfg.Webhook.Enable,
		SupportsUserVerification:   true,
		SupportsOnCallNotify:       true,
		SupportsSignals:            true,
		SupportsStatusUpdates:      true,
		SupportsAlertNotifications: true,
		StatusUpdatesRequired:      true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldWebhookURL,
			Label:              "Webhook URL",
			PlaceholderText:    "https://example.com",
			InputType:          "url",
			Hint:               "Webhook Documentation",
			HintURL:            "/docs#webhooks",
			SupportsValidation: true,
		}},
		DynamicParams: []nfydest.DynamicParamConfig{
			{
				ParamID: ParamBody,
				Label:   "Body",
				Hint:    "The body of the request.",
			},
			{
				ParamID:      ParamContentType,
				Label:        "Content Type",
				Hint:         "The content type (e.g., application/json).",
				DefaultValue: `"application/json"`, // Because this is an expression, it needs the double quotes.
			},
		},
	}, nil
}

func (CustomSender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeCustomWebhook,
		Name:                       "Custom Webhook",
		Enabled:                    cfg.CustomWebhook.Enable,
		SupportsUserVerification:   true,
		SupportsOnCallNotify:       true,
		SupportsStatusUpdates:      true,
		SupportsAlertNotifications: true,
		StatusUpdatesRequired:      true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldWebhookURL,
			Label:              "Webhook URL",
			PlaceholderText:    "https://example.com",
			InputType:          "url",
			Hint:               "Webhook Documentation",
			HintURL:            "/docs#webhooks",
			SupportsValidation: true,
		}, {
			FieldID:         FieldBodyTemplate,
			Label:           "Body Template",
			PlaceholderText: "{\"text\":\"{{.Summary}}\"}",
			InputType:       "text",
			Hint:            "Go template used as the request body. All notification fields are available.",
		}, {
			FieldID:         FieldContentType,
			Label:           "Content Type",
			PlaceholderText: "application/json",
			InputType:       "text",
			Hint:            "Optional. Defaults to application/json when left blank.",
		}},
	}, nil
}

func (s *Sender) ValidateField(ctx context.Context, fieldID, value string) error {
	cfg := config.FromContext(ctx)
	switch fieldID {
	case FieldWebhookURL:
		err := validate.AbsoluteURL(FieldWebhookURL, value)
		if err != nil {
			return err
		}
		if !cfg.ValidWebhookURL(value) {
			return validation.NewGenericError("url is not allowed by administator")
		}

		return nil
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *CustomSender) ValidateField(ctx context.Context, fieldID, value string) error {
	cfg := config.FromContext(ctx)
	switch fieldID {
	case FieldWebhookURL:
		err := validate.AbsoluteURL(FieldWebhookURL, value)
		if err != nil {
			return err
		}
		if !cfg.ValidWebhookURL(value) {
			return validation.NewGenericError("url is not allowed by administrator")
		}
		return nil
	case FieldBodyTemplate:
		if value == "" {
			return validation.NewFieldError(FieldBodyTemplate, "required")
		}
		_, err := parseTemplate(value)
		if err != nil {
			return err
		}
		return nil
	case FieldContentType:
		if value == "" {
			return nil
		}
		if _, _, err := mime.ParseMediaType(value); err != nil {
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

	u, err := url.Parse(args[FieldWebhookURL])
	if err != nil {
		return nil, validation.WrapError(err)
	}
	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Webhook",
		Text:        u.Hostname(),
	}, nil
}

func (s *CustomSender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	u, err := url.Parse(args[FieldWebhookURL])
	if err != nil {
		return nil, validation.WrapError(err)
	}
	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Custom Webhook",
		Text:        u.Hostname(),
	}, nil
}
