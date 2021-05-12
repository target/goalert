package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/target/goalert/notification"
)

type Sender struct{}

// POSTDataAlert represents fields in outgoing alert notification.
type POSTDataAlert struct {
	Type    string
	AlertID int
	Summary string
	Details string
}

// POSTDataAlertBundle represents fields in outgoing alert bundle notification.
type POSTDataAlertBundle struct {
	Type        string
	ServiceID   string
	ServiceName string
	Count       int
}

// POSTDataAlertStatus represents fields in outgoing alert status notification.
type POSTDataAlertStatus struct {
	Type     string
	AlertID  int
	LogEntry string
}

// POSTDataAlertStatusBundle represents fields in outgoing alert status bundle notification.
type POSTDataAlertStatusBundle struct {
	Type     string
	AlertID  int
	LogEntry string
	Count    int
}

// POSTDataVerification represents fields in outgoing verification notification.
type POSTDataVerification struct {
	Type string
	Code string
}

// POSTDataTest represents fields in outgoing test notification.
type POSTDataTest struct {
	Type string
}

func NewSender(ctx context.Context) *Sender {
	return &Sender{}
}

// Send will send an alert for the provided message type
func (s *Sender) Send(ctx context.Context, msg notification.Message) (*notification.MessageStatus, error) {

	var data []byte
	var err error

	switch m := msg.(type) {
	case notification.Test:
		pdTest := POSTDataTest{Type: "Test"}
		data, err = json.Marshal(pdTest)
	case notification.Verification:
		pdVerification := POSTDataVerification{Type: "Verification", Code: strconv.Itoa(m.Code)}
		data, err = json.Marshal(pdVerification)
	case notification.Alert:
		pdAlert := POSTDataAlert{Type: "Alert", Details: m.Details, AlertID: m.AlertID, Summary: m.Summary}
		data, err = json.Marshal(pdAlert)
	case notification.AlertBundle:
		pdAlertBundle := POSTDataAlertBundle{Type: "AlertBundle", ServiceID: m.ServiceID, ServiceName: m.ServiceName, Count: m.Count}
		data, err = json.Marshal(pdAlertBundle)
	case notification.AlertStatus:
		pdAlertStatus := POSTDataAlertStatus{Type: "AlertStatus", AlertID: m.AlertID, LogEntry: m.LogEntry}
		data, err = json.Marshal(pdAlertStatus)
	case notification.AlertStatusBundle:
		pdAlertStatusBundle := POSTDataAlertStatusBundle{Type: "AlertStatusBundle", Count: m.Count, AlertID: m.AlertID, LogEntry: m.LogEntry}
		data, err = json.Marshal(pdAlertStatusBundle)
	default:
		return nil, fmt.Errorf("message type: %d not supported", m.Type())
	}

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", msg.Destination().Value, bytes.NewReader(data))
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
