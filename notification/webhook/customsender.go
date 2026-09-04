package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/validation"
)

type CustomSender struct {
	Client *http.Client
}

type customWebhookPayload struct {
	MessageID string
	AppName   string
	Type      string

	AlertID       int
	Summary       string
	Details       string
	ServiceID     string
	ServiceName   string
	Count         int
	LogEntry      string
	Code          string
	ScheduleID    string
	ScheduleName  string
	ScheduleURL   string
	Users         []notification.User
	Meta          map[string]string
	NewAlertState string
}

func (s *CustomSender) client() *http.Client {
	if s != nil && s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

func (s *CustomSender) SendMessage(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	cfg := config.FromContext(ctx)

	webURL := msg.DestArg(FieldWebhookURL)
	if !cfg.ValidWebhookURL(webURL) {
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: "invalid or not allowed URL",
		}, nil
	}

	tpl, err := template.New("body").Option("missingkey=error").Parse(msg.DestArg(FieldBodyTemplate))
	if err != nil {
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: err.Error(),
		}, nil
	}

	data := renderCustomWebhookPayload(cfg, msg)
	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: err.Error(),
		}, nil
	}

	contentType := msg.DestArg(FieldContentType)
	if contentType == "" {
		contentType = "application/json"
	}
	if _, _, err := mime.ParseMediaType(contentType); err != nil {
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: err.Error(),
		}, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webURL, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	resp, err := s.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return &notification.SentMessage{State: notification.StateSent}, nil
	}

	respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if readErr != nil {
		return nil, readErr
	}

	details := strings.TrimSpace(string(respBody))
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

func renderCustomWebhookPayload(cfg config.Config, msg notification.Message) customWebhookPayload {
	res := customWebhookPayload{
		MessageID: msg.MsgID(),
		AppName:   cfg.ApplicationName(),
		Type:      fmt.Sprintf("%T", msg),
	}

	switch t := msg.(type) {
	case notification.Test:
		res.Type = "Test"
	case notification.Verification:
		res.Type = "Verification"
		res.Code = t.Code
	case notification.Alert:
		res.Type = "Alert"
		res.AlertID = t.AlertID
		res.Summary = t.Summary
		res.Details = t.Details
		res.ServiceID = t.ServiceID
		res.ServiceName = t.ServiceName
		res.Meta = t.Meta
	case notification.AlertBundle:
		res.Type = "AlertBundle"
		res.ServiceID = t.ServiceID
		res.ServiceName = t.ServiceName
		res.Count = t.Count
	case notification.AlertStatus:
		res.Type = "AlertStatus"
		res.AlertID = t.AlertID
		res.LogEntry = t.LogEntry
		res.ServiceID = t.ServiceID
		res.Summary = t.Summary
		res.Details = t.Details
		res.NewAlertState = alertStateText(t.NewAlertState)
	case notification.ScheduleOnCallUsers:
		res.Type = "ScheduleOnCallUsers"
		res.ScheduleID = t.ScheduleID
		res.ScheduleName = t.ScheduleName
		res.ScheduleURL = t.ScheduleURL
		res.Users = append([]notification.User(nil), t.Users...)
	default:
		res.Type = fmt.Sprintf("%T", msg)
	}

	return res
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

func parseTemplate(body string) (*template.Template, error) {
	if body == "" {
		return nil, validation.NewFieldError(FieldBodyTemplate, "required")
	}
	tpl, err := template.New("body").Option("missingkey=error").Parse(body)
	if err != nil {
		return nil, validation.WrapError(err)
	}
	return tpl, nil
}
