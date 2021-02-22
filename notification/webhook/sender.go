package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/target/goalert/notification"
)

type Sender struct{}

// POSTDataType is the type of alert data being sent
type POSTDataType string

const (
	TypeTest              POSTDataType = "Test"
	TypeVerification      POSTDataType = "Verification"
	TypeAlert             POSTDataType = "Alert"
	TypeAlertBundle       POSTDataType = "AlertBundle"
	TypeAlertStatus       POSTDataType = "AlertStatus"
	TypeAlertStatusBundle POSTDataType = "AlertStatusBundle"
)

// POSTData is a union of all possible message types, should be populated accordingly
type POSTData struct {
	AlertID     int          `json:",omitempty"`
	Type        POSTDataType `json:",omitempty"`
	Code        string       `json:",omitempty"`
	Summary     string       `json:",omitempty"`
	Details     string       `json:",omitempty"`
	ServiceID   string       `json:",omitempty"`
	ServiceName string       `json:",omitempty"`
	Count       int          `json:",omitempty"`
	LogEntry    string       `json:",omitempty"`
}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

// Send will send an alert for the provided message type
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {

	var payload *POSTData

	switch m := msg.(type) {
	case notification.Test:
		payload = &POSTData{
			Type: TypeTest,
		}
	case notification.Verification:
		payload = &POSTData{
			Type: TypeVerification,
			Code: strconv.Itoa(m.Code),
		}
	case notification.Alert:
		payload = &POSTData{
			Type:    TypeAlert,
			AlertID: m.AlertID,
			Summary: m.Summary,
			Details: m.Details,
		}
	case notification.AlertBundle:
		payload = &POSTData{
			Type:        TypeAlertBundle,
			ServiceID:   m.ServiceID,
			ServiceName: m.ServiceName,
			Count:       m.Count,
		}
	case notification.AlertStatus:
		payload = &POSTData{
			Type:     TypeAlertStatus,
			AlertID:  m.AlertID,
			LogEntry: m.LogEntry,
		}
	case notification.AlertStatusBundle:
		payload = &POSTData{
			Type:     TypeAlertStatusBundle,
			Count:    m.Count,
			AlertID:  m.AlertID,
			LogEntry: m.LogEntry,
		}
	default:
		return nil, errors.New("message type: " + string(m.Type()) + " not supported")
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", msg.Destination().Value, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &notification.MessageStatus{ID: msg.ID(), State: notification.MessageStateSent, ProviderMessageID: msg.ID()}, nil
}

func (s *Sender) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	return nil, errors.New("notification/webhook: status not supported")
}

func (s *Sender) ListenStatus() <-chan *notification.MessageStatus     { return nil }
func (s *Sender) ListenResponse() <-chan *notification.MessageResponse { return nil }
