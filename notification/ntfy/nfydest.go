package ntfy

import (
	"context"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const (
	DestTypeNtfy   = "builtin-ntfy"
	FieldNtfyTopic = "ntfy_topic"

	// FallbackIconURL reuses the webhook icon. The UI resolves builtin icons from a fixed map and
	// renders anything else as an image URL, so an unknown builtin value shows as a broken image.
	FallbackIconURL = "builtin://webhook"

	// maxTopicLen is ntfy's limit on a topic name.
	maxTopicLen = 64
)

// NewNtfyDest returns a destination for the given ntfy topic.
func NewNtfyDest(topic string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeNtfy, FieldNtfyTopic, topic)
}

var _ (nfydest.Provider) = (*Sender)(nil)

func (Sender) ID() string { return DestTypeNtfy }

func (Sender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeNtfy,
		Name:                       "Ntfy",
		IconURL:                    FallbackIconURL,
		IconAltText:                "Ntfy",
		Enabled:                    cfg.Ntfy.Enable,
		SupportsUserVerification:   true,
		SupportsOnCallNotify:       true,
		SupportsStatusUpdates:      true,
		SupportsAlertNotifications: true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldNtfyTopic,
			Label:              "Topic",
			PlaceholderText:    "alerts-jane",
			InputType:          "text",
			Hint:               "Subscribe to this topic in the ntfy app to receive notifications.",
			SupportsValidation: true,
		}},
	}, nil
}

func (s *Sender) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldNtfyTopic:
		return validateTopic(value)
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *Sender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	topic := args[FieldNtfyTopic]
	if err := validateTopic(topic); err != nil {
		return nil, validation.WrapError(err)
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Ntfy",
		Text:        topic,
		LinkURL:     topicURL(config.FromContext(ctx).Ntfy.ServerURL, topic),
	}, nil
}

// validateTopic enforces ntfy's topic charset. A topic is a URL path segment on the ntfy server, so
// anything outside this set either changes the path or is rejected on publish.
func validateTopic(topic string) error {
	if err := validate.ASCII(FieldNtfyTopic, topic, 1, maxTopicLen); err != nil {
		return err
	}

	for _, r := range topic {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return validation.NewFieldError(FieldNtfyTopic,
				"may only contain letters, numbers, dashes and underscores")
		}
	}

	return nil
}

// topicURL is the publish endpoint for a topic, and also the page a browser can subscribe from.
func topicURL(serverURL, topic string) string {
	if serverURL == "" || topic == "" {
		return ""
	}

	return strings.TrimSuffix(serverURL, "/") + "/" + topic
}
