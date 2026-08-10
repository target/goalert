package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/retry"
)

type Sender struct {
	Client *http.Client
}

const idempotencyKeyHeader = "Idempotency-Key"

const (
	alertStateUnacknowledged = "Unacknowledged"
	alertStateAcknowledged   = "Acknowledged"
	alertStateClosed         = "Closed"
)

// POSTDataAlert represents fields in outgoing alert notification.
type POSTDataAlert struct {
	AppName     string
	Type        string
	AlertID     int
	Summary     string
	Details     string
	ServiceID   string
	ServiceName string
	Meta        map[string]string
}

// POSTDataAlertBundle represents fields in outgoing alert bundle notification.
type POSTDataAlertBundle struct {
	AppName     string
	Type        string
	ServiceID   string
	ServiceName string
	Count       int
}

// POSTDataAlertStatus represents fields in outgoing alert status notification.
type POSTDataAlertStatus struct {
	AppName    string
	Type       string
	AlertID    int
	LogEntry   string
	AlertState string
}

// POSTDataAlertStatusBundle represents fields in outgoing alert status bundle notification.
type POSTDataAlertStatusBundle struct {
	AppName  string
	Type     string
	AlertID  int
	LogEntry string
	Count    int
}

// POSTDataVerification represents fields in outgoing verification notification.
type POSTDataVerification struct {
	AppName string
	Type    string
	Code    string
}

// POSTDataOnCallUser represents User fields in outgoing on call notification.
type POSTDataOnCallUser struct {
	ID   string
	Name string
	URL  string
}

// POSTDataOnCallNotification represents fields in outgoing on call notification.
type POSTDataOnCallNotification struct {
	AppName      string
	Type         string
	Users        []POSTDataOnCallUser
	ScheduleID   string
	ScheduleName string
	ScheduleURL  string
}

// POSTDataTest represents fields in outgoing test notification.
type POSTDataTest struct {
	AppName string
	Type    string
}

func NewSender(ctx context.Context, client *http.Client) *Sender {
	return &Sender{
		Client: client,
	}
}

var _ nfydest.MessageSender = &Sender{}

func safeRequestError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return fmt.Errorf("webhook request failed: %w", context.Canceled)
	case errors.Is(err, context.DeadlineExceeded):
		return fmt.Errorf("webhook request failed: %w", context.DeadlineExceeded)
	default:
		return retry.TemporaryError(errors.New("webhook request failed"))
	}
}

func alertStateWireValue(state notification.AlertState) (string, error) {
	switch state {
	case notification.AlertStateUnacknowledged:
		return alertStateUnacknowledged, nil
	case notification.AlertStateAcknowledged:
		return alertStateAcknowledged, nil
	case notification.AlertStateClosed:
		return alertStateClosed, nil
	default:
		return "", errors.New("webhook alert state is invalid")
	}
}

// Send will send an alert for the provided message type
func (s *Sender) SendMessage(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
	deliveryID := msg.MsgID()
	if deliveryID == "" {
		return nil, errors.New("webhook delivery identity is required")
	}

	cfg := config.FromContext(ctx)
	var payload interface{}
	switch m := msg.(type) {
	case notification.Test:
		payload = POSTDataTest{
			AppName: cfg.ApplicationName(),
			Type:    "Test",
		}
	case notification.Verification:
		payload = POSTDataVerification{
			AppName: cfg.ApplicationName(),
			Type:    "Verification",
			Code:    m.Code,
		}
	case notification.Alert:
		payload = POSTDataAlert{
			AppName:     cfg.ApplicationName(),
			Type:        "Alert",
			Details:     m.Details,
			AlertID:     m.AlertID,
			Summary:     m.Summary,
			ServiceID:   m.ServiceID,
			ServiceName: m.ServiceName,
			Meta:        m.Meta,
		}
	case notification.AlertBundle:
		payload = POSTDataAlertBundle{
			AppName:     cfg.ApplicationName(),
			Type:        "AlertBundle",
			ServiceID:   m.ServiceID,
			ServiceName: m.ServiceName,
			Count:       m.Count,
		}
	case notification.AlertStatus:
		alertState, err := alertStateWireValue(m.NewAlertState)
		if err != nil {
			return nil, err
		}
		payload = POSTDataAlertStatus{
			AppName:    cfg.ApplicationName(),
			Type:       "AlertStatus",
			AlertID:    m.AlertID,
			LogEntry:   m.LogEntry,
			AlertState: alertState,
		}
	case notification.ScheduleOnCallUsers:
		// We use types defined in this package to insulate against unintended API
		// changes.
		users := make([]POSTDataOnCallUser, len(m.Users))
		for i, u := range m.Users {
			users[i] = POSTDataOnCallUser(u)
		}
		payload = POSTDataOnCallNotification{
			AppName:      cfg.ApplicationName(),
			Type:         "ScheduleOnCallUsers",
			Users:        users,
			ScheduleID:   m.ScheduleID,
			ScheduleName: m.ScheduleName,
			ScheduleURL:  m.ScheduleURL,
		}
	default:
		return nil, fmt.Errorf("message type '%T' not supported", m)
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	webURL := msg.DestArg(FieldWebhookURL)
	if !cfg.ValidWebhookURL(webURL) {
		// fail permanently if the URL is not currently valid/allowed
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: "invalid or not allowed URL",
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webURL, bytes.NewReader(data))
	if err != nil {
		return nil, errors.New("webhook request could not be created")
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set(idempotencyKeyHeader, deliveryID)

	resp, err := s.Client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, safeRequestError(err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("webhook request failed: HTTP status %d", resp.StatusCode)
	}

	return &notification.SentMessage{State: notification.StateSent}, nil
}
