package teams

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
)

// Sender posts Adaptive Card messages to Microsoft Teams channels using
// Power Automate Workflow webhook URLs.
type Sender struct {
	client *http.Client
}

// NewSender creates a new Sender using the given HTTP client.
func NewSender(ctx context.Context, client *http.Client) *Sender {
	if client == nil {
		client = http.DefaultClient
	}
	return &Sender{client: client}
}

var _ nfydest.MessageSender = &Sender{}

// workflowMessage is the payload format expected by the Teams
// "post a card in a chat or channel" Workflow trigger.
type workflowMessage struct {
	Type        string       `json:"type"`
	Attachments []attachment `json:"attachments"`
}

type attachment struct {
	ContentType string       `json:"contentType"`
	ContentURL  *string      `json:"contentUrl"`
	Content     AdaptiveCard `json:"content"`
}

// NewWorkflowMessage wraps an Adaptive Card in the message envelope expected
// by a Teams Workflow webhook.
func NewWorkflowMessage(card AdaptiveCard) any {
	return workflowMessage{
		Type: "message",
		Attachments: []attachment{{
			ContentType: "application/vnd.microsoft.card.adaptive",
			Content:     card,
		}},
	}
}

// SendMessage renders the message as an Adaptive Card and posts it to the
// destination's workflow webhook URL.
func (s *Sender) SendMessage(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)

	var card AdaptiveCard
	switch m := msg.(type) {
	case notification.Test:
		card = testCard(ctx)
	case notification.Alert:
		// Workflow webhooks cannot update existing messages, so
		// re-notifications post a fresh card.
		card = alertCard(ctx, m.AlertID, m.Summary, m.Details, m.ServiceName, "", notification.AlertStateUnacknowledged)
	case notification.AlertStatus:
		card = alertCard(ctx, m.AlertID, m.Summary, m.Details, m.ServiceName, m.LogEntry, m.NewAlertState)
	case notification.AlertBundle:
		card = alertBundleCard(ctx, m.ServiceID, m.ServiceName, m.Count)
	case notification.SignalMessage:
		card = signalCard(ctx, m.Param(ParamMessage))
	case notification.ScheduleOnCallUsers:
		card = onCallCard(ctx, m)
	default:
		return nil, fmt.Errorf("message type '%T' not supported", m)
	}

	webURL := msg.DestArg(FieldWebhookURL)
	if !cfg.ValidTeamsWorkflowURL(webURL) {
		// fail permanently if the URL is not currently valid/allowed
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: "invalid or not allowed URL",
		}, nil
	}

	data, err := json.Marshal(NewWorkflowMessage(card))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", webURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	switch {
	case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
		// temporary failure; return an error so the message is retried
		return nil, fmt.Errorf("post to teams workflow: %s", resp.Status)
	case resp.StatusCode >= 400:
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: fmt.Sprintf("teams workflow responded with %s", resp.Status),
		}, nil
	}

	return &notification.SentMessage{State: notification.StateSent}, nil
}
