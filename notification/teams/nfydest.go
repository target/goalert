package teams

import (
	"context"
	"net/url"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const (
	DestTypeTeamsWorkflow = "builtin-teams-workflow"
	FieldWebhookURL       = "teams_webhook_url"
	ParamMessage          = "message"
	FallbackIconURL       = "builtin://msteams"
)

// NewTeamsWorkflowDest creates a new Teams workflow webhook destination.
func NewTeamsWorkflowDest(url string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeTeamsWorkflow, FieldWebhookURL, url)
}

var _ nfydest.Provider = (*Sender)(nil)

func (Sender) ID() string { return DestTypeTeamsWorkflow }

func (Sender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeTeamsWorkflow,
		Name:                       "Microsoft Teams",
		Enabled:                    cfg.Teams.Enable,
		SupportsAlertNotifications: true,
		SupportsStatusUpdates:      true,
		SupportsOnCallNotify:       true,
		SupportsSignals:            true,
		StatusUpdatesRequired:      true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldWebhookURL,
			Label:              "Workflow Webhook URL",
			PlaceholderText:    "https://example.westus.logic.azure.com/workflows/...",
			InputType:          "url",
			Hint:               "Teams Documentation",
			HintURL:            "/docs#teams",
			SupportsValidation: true,
		}},
		DynamicParams: []nfydest.DynamicParamConfig{{
			ParamID: ParamMessage,
			Label:   "Message",
			Hint:    "The text of the message to send.",
		}},
	}, nil
}

func validateWebhookURL(ctx context.Context, value string) error {
	cfg := config.FromContext(ctx)

	err := validate.AbsoluteURL(FieldWebhookURL, value)
	if err != nil {
		return err
	}

	u, err := url.Parse(value)
	if err != nil {
		return validation.WrapError(err)
	}
	if u.Scheme != "https" {
		return validation.NewGenericError("url must use https")
	}
	if !cfg.ValidTeamsWorkflowURL(value) {
		return validation.NewGenericError("url is not allowed by administrator")
	}

	return nil
}

func (s *Sender) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldWebhookURL:
		return validateWebhookURL(ctx, value)
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
		IconAltText: "Microsoft Teams",
		Text:        u.Hostname(),
	}, nil
}
