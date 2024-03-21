package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

type Sender struct{}

// POSTDataAlert represents fields in outgoing alert notification.
type POSTDataAlert struct {
	AppName     string
	Type        string
	AlertID     int
	Summary     string
	Details     string
	ServiceID   string
	ServiceName string
	Meta        []notification.AlertMeta
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
	AppName  string
	Type     string
	AlertID  int
	LogEntry string
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

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

// Send will send an alert for the provided message type
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.SentMessage, error) {
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
			Code:    strconv.Itoa(m.Code),
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
		payload = POSTDataAlertStatus{
			AppName:  cfg.ApplicationName(),
			Type:     "AlertStatus",
			AlertID:  m.AlertID,
			LogEntry: m.LogEntry,
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
		return nil, fmt.Errorf("message type '%s' not supported", m.Type().String())
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if !cfg.ValidWebhookURL(msg.Destination().Value) {
		// fail permanently if the URL is not currently valid/allowed
		return &notification.SentMessage{
			State:        notification.StateFailedPerm,
			StateDetails: "invalid or not allowed URL",
		}, nil
	}

	req, err := http.NewRequestWithContext(ctx, "POST", msg.Destination().Value, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &notification.SentMessage{State: notification.StateSent}, nil
}
