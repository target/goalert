package googlechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

const (
	DestTypeGoogleChat = "builtin-google-chat"
	FieldWebhookURL    = "webhook_url"

	FallbackIconURL = "builtin://googlechat"
)

var googleChatWebhookPath = regexp.MustCompile(`^/v1/spaces/[^/]+/messages$`)

type Sender struct {
	Client *http.Client
}

type ChatMessage struct {
	Text string `json:"text"`
}

var _ nfydest.Provider = (*Sender)(nil)
var _ nfydest.MessageSender = (*Sender)(nil)

func NewSender(_ context.Context, client *http.Client) *Sender {
	return &Sender{Client: client}
}

func NewGoogleChatDest(webhookURL string) gadb.DestV1 {
	return gadb.NewDestV1(DestTypeGoogleChat, FieldWebhookURL, webhookURL)
}

func (Sender) ID() string { return DestTypeGoogleChat }

func (Sender) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeGoogleChat,
		Name:                       "Google Chat",
		Enabled:                    cfg.GoogleChat.Enable,
		IconURL:                    FallbackIconURL,
		IconAltText:                "Google Chat",
		UserDisclaimer:             "Google Chat webhooks support alert and on-call change notifications.",
		SupportsAlertNotifications: true,
		SupportsStatusUpdates:      true,
		SupportsOnCallNotify:       true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:            FieldWebhookURL,
			Label:              "Webhook URL",
			PlaceholderText:    "https://chat.googleapis.com/v1/spaces/.../messages?key=...&token=...",
			InputType:          "url",
			SupportsValidation: true,
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

		if err := validateGoogleChatWebhookURL(value); err != nil {
			return err
		}

		if !cfg.ValidWebhookURL(value) {
			return validation.NewGenericError("url is not allowed by administrator")
		}

		return nil
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *Sender) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	if err := validateGoogleChatWebhookURL(args[FieldWebhookURL]); err != nil {
		return nil, err
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Google Chat",
		Text:        "Google Chat",
	}, nil
}

func (s *Sender) SendMessage(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)
	webhookURL := msg.DestArg(FieldWebhookURL)

	if err := validateGoogleChatWebhookURL(webhookURL); err != nil {
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: err.Error(),
		}, nil
	}

	if !cfg.ValidWebhookURL(webhookURL) {
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: "invalid or not allowed URL",
		}, nil
	}

	payload, err := json.Marshal(ChatMessage{Text: formatGoAlertMessage(ctx, msg)})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := s.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &notification.SentMessage{State: notification.StateSent}, nil
	}

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if readErr != nil {
		return nil, readErr
	}

	details := strings.TrimSpace(string(body))
	if details != "" {
		details = fmt.Sprintf("%s: %s", resp.Status, details)
	} else {
		details = resp.Status
	}

	state := notification.StateFailedPerm
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		state = notification.StateFailedTemp
	}

	return &notification.SentMessage{
		State:        state,
		StateDetails: details,
	}, nil
}

func (s *Sender) client() *http.Client {
	if s != nil && s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

func formatGoAlertMessage(ctx context.Context, msg notification.Message) string {
	cfg := config.FromContext(ctx)

	switch t := msg.(type) {
	case notification.Test:
		return "This is a test message from GoAlert."
	case notification.Alert:
		return formatAlertMessage(cfg, t)
	case notification.AlertBundle:
		return formatAlertBundleMessage(cfg, t)
	case notification.AlertStatus:
		return formatAlertStatusMessage(cfg, t)
	case notification.ScheduleOnCallUsers:
		return formatScheduleOnCallUsers(t)
	default:
		return fmt.Sprintf("unsupported message type: %T", t)
	}
}

func formatAlertMessage(cfg config.Config, a notification.Alert) string {
	var parts []string
	parts = append(parts,
		"GoAlert alert",
		fmt.Sprintf("Alert: #%d %s", a.AlertID, a.Summary),
		fmt.Sprintf("Service: %s", a.ServiceName),
	)
	if a.Details != "" {
		parts = append(parts, fmt.Sprintf("Details: %s", a.Details))
	}
	parts = append(parts, fmt.Sprintf("Link: %s", alertURL(cfg, a.AlertID)))

	return strings.Join(parts, "\n")
}

func formatAlertBundleMessage(cfg config.Config, a notification.AlertBundle) string {
	return strings.Join([]string{
		"GoAlert alert bundle",
		fmt.Sprintf("Service: %s", a.ServiceName),
		fmt.Sprintf("Count: %d unacknowledged alerts", a.Count),
		fmt.Sprintf("Link: %s", serviceAlertsURL(cfg, a.ServiceID)),
	}, "\n")
}

func formatAlertStatusMessage(cfg config.Config, a notification.AlertStatus) string {
	state := alertStateText(a.NewAlertState)
	if state == "" {
		state = "updated"
	}

	var parts []string
	parts = append(parts,
		"GoAlert alert update",
		fmt.Sprintf("Alert: #%d %s", a.AlertID, a.Summary),
		fmt.Sprintf("State: %s", state),
	)
	if a.LogEntry != "" {
		parts = append(parts, fmt.Sprintf("Log: %s", a.LogEntry))
	}
	parts = append(parts, fmt.Sprintf("Link: %s", alertURL(cfg, a.AlertID)))

	return strings.Join(parts, "\n")
}

func alertStateText(state notification.AlertState) string {
	switch state {
	case notification.AlertStateUnacknowledged:
		return "unacknowledged"
	case notification.AlertStateAcknowledged:
		return "acknowledged"
	case notification.AlertStateClosed:
		return "closed"
	default:
		return ""
	}
}

func alertURL(cfg config.Config, id int) string {
	return cfg.CallbackURL(fmt.Sprintf("/alerts/%d", id))
}

func serviceAlertsURL(cfg config.Config, serviceID string) string {
	return cfg.CallbackURL(fmt.Sprintf("/services/%s/alerts", serviceID))
}

func formatScheduleOnCallUsers(n notification.ScheduleOnCallUsers) string {
	users := "Nobody"
	if len(n.Users) > 0 {
		sorted := make([]notification.User, len(n.Users))
		copy(sorted, n.Users)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].Name == sorted[j].Name {
				return sorted[i].ID < sorted[j].ID
			}
			return sorted[i].Name < sorted[j].Name
		})

		names := make([]string, 0, len(sorted))
		for _, u := range sorted {
			names = append(names, u.Name)
		}
		users = strings.Join(names, ", ")
	}

	return fmt.Sprintf(
		"GoAlert on-call shift changed\nSchedule: %s\nNow on-call: %s\nLink: %s",
		n.ScheduleName,
		users,
		n.ScheduleURL,
	)
}

func validateGoogleChatWebhookURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return validation.WrapError(err)
	}

	if u.Scheme != "https" {
		return validation.NewFieldError(FieldWebhookURL, "must use https")
	}
	if u.Hostname() != "chat.googleapis.com" {
		return validation.NewFieldError(FieldWebhookURL, "must be a Google Chat incoming webhook URL")
	}
	if !googleChatWebhookPath.MatchString(u.Path) {
		return validation.NewFieldError(FieldWebhookURL, "must be a Google Chat incoming webhook URL")
	}
	q := u.Query()
	if q.Get("key") == "" || q.Get("token") == "" {
		return validation.NewFieldError(FieldWebhookURL, "must include key and token query parameters")
	}

	return nil
}
