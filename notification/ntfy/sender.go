package ntfy

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/notification/nfymsg"
)

// requestTimeout matches the webhook sender: a page that has not left the building within a few
// seconds is better retried than held.
const requestTimeout = 3 * time.Second

// ntfy priority levels. Alerts publish at max so they break through Do Not Disturb on the phone;
// everything else is informational.
const (
	priorityDefault = 3
	priorityHigh    = 4
	priorityMax     = 5
)

// MetaClickURL is the alert metadata key whose value becomes the notification's tap target. A caller
// that creates alerts with a link into its own system -- a chat thread, a runbook -- sets it there,
// because the payload GoAlert builds has no other way to carry one.
const MetaClickURL = "click"

// maxTitleLen caps the title, which travels as an HTTP header rather than in the body.
const maxTitleLen = 250

type Sender struct {
	Client *http.Client
}

func NewSender(ctx context.Context, client *http.Client) *Sender {
	return &Sender{Client: client}
}

var _ nfydest.MessageSender = &Sender{}

// message is one ntfy publish: a body, plus the headers ntfy renders it with.
type message struct {
	Title    string
	Body     string
	Priority int
	ClickURL string
	Tags     string
}

// SendMessage publishes a notification to the destination's ntfy topic.
func (s *Sender) SendMessage(ctx context.Context, msg nfymsg.Message) (*nfymsg.SentMessage, error) {
	cfg := config.FromContext(ctx)

	topic := msg.DestArg(FieldNtfyTopic)
	if err := validateTopic(topic); err != nil {
		return &nfymsg.SentMessage{
			State:        nfymsg.StateFailedPerm,
			StateDetails: "invalid ntfy topic",
		}, nil
	}

	endpoint := topicURL(cfg.Ntfy.ServerURL, topic)
	if endpoint == "" {
		return &nfymsg.SentMessage{
			State:        nfymsg.StateFailedPerm,
			StateDetails: "no ntfy server URL configured",
		}, nil
	}

	m, err := buildMessage(cfg, msg)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(m.Body))
	if err != nil {
		return nil, err
	}

	if title := headerValue(m.Title, maxTitleLen); title != "" {
		req.Header.Set("X-Title", title)
	}
	req.Header.Set("X-Priority", strconv.Itoa(m.Priority))
	if cfg.Ntfy.Markdown {
		req.Header.Set("X-Markdown", "yes")
	}
	if m.ClickURL != "" {
		req.Header.Set("X-Click", m.ClickURL)
	}
	if m.Tags != "" {
		req.Header.Set("X-Tags", m.Tags)
	}
	if cfg.Ntfy.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Ntfy.Token)
	}

	client := s.Client
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// A 4xx fails the same way on a retry: a rejected token, or a topic the server will not accept.
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return &nfymsg.SentMessage{
			State:        nfymsg.StateFailedPerm,
			StateDetails: fmt.Sprintf("ntfy returned %d", resp.StatusCode),
		}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ntfy returned %d", resp.StatusCode)
	}

	return &nfymsg.SentMessage{State: nfymsg.StateSent}, nil
}

// buildMessage renders a notification into the fields ntfy publishes.
func buildMessage(cfg config.Config, msg nfymsg.Message) (message, error) {
	switch m := msg.(type) {
	case nfymsg.Test:
		return message{
			Title:    cfg.ApplicationName(),
			Body:     "Test message.",
			Priority: priorityDefault,
		}, nil

	case nfymsg.Verification:
		return message{
			Title:    cfg.ApplicationName(),
			Body:     fmt.Sprintf("Verification code: %s", m.Code),
			Priority: priorityHigh,
			Tags:     "key",
		}, nil

	case nfymsg.Alert:
		return message{
			Title:    m.Summary,
			Body:     m.Details,
			Priority: priorityMax,
			ClickURL: alertClickURL(cfg, m.Meta, m.AlertID),
			Tags:     "rotating_light",
		}, nil

	case nfymsg.AlertBundle:
		return message{
			Title:    m.ServiceName,
			Body:     fmt.Sprintf("%d unacknowledged alerts.", m.Count),
			Priority: priorityMax,
			ClickURL: cfg.CallbackURL(fmt.Sprintf("/services/%s/alerts", m.ServiceID)),
			Tags:     "rotating_light",
		}, nil

	case nfymsg.AlertStatus:
		return message{
			Title:    alertStatusTitle(m),
			Body:     m.LogEntry,
			Priority: priorityDefault,
			ClickURL: alertClickURL(cfg, nil, m.AlertID),
		}, nil

	case nfymsg.ScheduleOnCallUsers:
		return message{
			Title:    fmt.Sprintf("On call for %s", m.ScheduleName),
			Body:     onCallBody(m.Users),
			Priority: priorityDefault,
			ClickURL: m.ScheduleURL,
			Tags:     "calendar",
		}, nil
	}

	return message{}, fmt.Errorf("message type '%T' not supported", msg)
}

// alertClickURL is where tapping the notification lands. Alert metadata wins, so the system that
// raised the alert can point at itself; otherwise the alert in GoAlert.
func alertClickURL(cfg config.Config, meta map[string]string, alertID int) string {
	if u, ok := meta[MetaClickURL]; ok && isAbsoluteURL(u) {
		return u
	}

	return cfg.CallbackURL(fmt.Sprintf("/alerts/%d", alertID))
}

// alertStatusTitle falls back to the alert number, because the summary is not always carried on a
// status update.
func alertStatusTitle(m nfymsg.AlertStatus) string {
	if m.Summary != "" {
		return m.Summary
	}

	return fmt.Sprintf("Alert #%d", m.AlertID)
}

func onCallBody(users []nfymsg.User) string {
	if len(users) == 0 {
		return "Nobody is on call."
	}

	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Name
	}

	return strings.Join(names, ", ")
}

func isAbsoluteURL(s string) bool {
	if s == "" {
		return false
	}

	u, err := url.Parse(s)

	return err == nil && u.IsAbs()
}

// headerValue flattens a value for use as an HTTP header. ntfy takes the title as a header, and a
// newline in an alert summary would otherwise make the request unsendable.
func headerValue(s string, maxLen int) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}

		return r
	}, s)

	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		s = strings.TrimSpace(s[:maxLen])
	}

	return s
}
