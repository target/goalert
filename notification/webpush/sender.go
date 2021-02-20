package webpush

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

type Sender struct{}

// WebPushDataType is the type of alert data being sent
type WebPushDataType string

const (
	TypeTest              WebPushDataType = "Test"
	TypeVerification      WebPushDataType = "Verification"
	TypeAlert             WebPushDataType = "Alert"
	TypeAlertBundle       WebPushDataType = "AlertBundle"
	TypeAlertStatus       WebPushDataType = "AlertStatus"
	TypeAlertStatusBundle WebPushDataType = "AlertStatusBundle"
)

// POSTData is a union of all possible message types, should be populated accordingly
type POSTData struct {
	Message string          `json:",omitempty"`
	Type    WebPushDataType `json:",omitempty"`
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
			Type:    TypeTest,
			Message: "GoAlert test",
		}
	case notification.Verification:
		payload = &POSTData{
			Type:    TypeVerification,
			Message: "GoAlert verify code: " + strconv.Itoa(m.Code),
		}
	case notification.Alert:
		payload = &POSTData{
			Type:    TypeAlert,
			Message: m.Summary,
		}
	case notification.AlertBundle:
		payload = &POSTData{
			Type:    TypeAlertBundle,
			Message: m.ServiceName + " has an alert bundle",
		}
	case notification.AlertStatus:
		payload = &POSTData{
			Type:    TypeAlertStatus,
			Message: "Alert Status Update for Alert #" + strconv.Itoa(m.AlertID),
		}
	case notification.AlertStatusBundle:
		payload = &POSTData{
			Type:    TypeAlertStatusBundle,
			Message: "Alert Status Update Bundle for " + strconv.Itoa(m.Count) + " alerts",
		}

	default:
		return nil, errors.New("message type not supported")
	}

	cfg := config.FromContext(ctx)

	sub := webpush.Subscription{}
	err := json.Unmarshal([]byte(msg.Destination().Value), &sub)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(payload)

	data = []byte(`{"url":"https://www.google.com", "message":"` + payload.Message + `"}`)

	// Send Notification
	resp, err := webpush.SendNotification(data, &sub, &webpush.Options{
		VAPIDPublicKey:  cfg.WebPushNotifications.VAPIDPublicKey,
		VAPIDPrivateKey: cfg.WebPushNotifications.VAPIDPrivateKey,
		TTL:             3600,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &notification.MessageStatus{ID: msg.ID(), State: notification.MessageStateSent, ProviderMessageID: msg.ID()}, nil
}

func (s *Sender) Status(ctx context.Context, id, providerID string) (*notification.MessageStatus, error) {
	return nil, errors.New("notification/webhook: status not supported")
}

func (s *Sender) ListenStatus() <-chan *notification.MessageStatus     { return nil }
func (s *Sender) ListenResponse() <-chan *notification.MessageResponse { return nil }
