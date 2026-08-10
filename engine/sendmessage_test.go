package engine

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/engine/message"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/notification/nfymsg"
	"github.com/target/goalert/notification/webhook"
	"github.com/target/goalert/retry"
	"github.com/target/goalert/util/log"
)

type deliveryRoundTripFunc func(*http.Request) (*http.Response, error)

func (f deliveryRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSendMessagePropagatesWebhookDeliveryIdentity(t *testing.T) {
	const (
		firstOutgoingMessageID  = "11111111-2222-4333-8444-555555555555"
		secondOutgoingMessageID = "aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee"
		webURL                  = "https://gateway.invalid/v1/goalert/contact-method/opaque-secret-token?route=secret-query"
		wantBody                = `{"AppName":"GoAlert","Type":"Test"}`
	)

	type capturedRequest struct {
		deliveryID string
		body       string
	}
	var requests []capturedRequest
	client := &http.Client{
		Transport: deliveryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			requests = append(requests, capturedRequest{
				deliveryID: req.Header.Get("Idempotency-Key"),
				body:       string(body),
			})
			if len(requests) == 1 {
				return nil, errors.New("dial failed for " + webURL + ": " + wantBody)
			}
			return &http.Response{
				StatusCode: http.StatusAccepted,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader("sensitive-response-content")),
				Request:    req,
			}, nil
		}),
	}

	var logOutput bytes.Buffer
	logger := log.NewLogger()
	logger.EnableDebug()
	logger.SetOutput(&logOutput)
	var appConfig config.Config
	appConfig.Webhook.Enable = true
	ctx := appConfig.Context(logger.BackgroundContext())
	registry := nfydest.NewRegistry()
	registry.RegisterProvider(ctx, webhook.NewSender(ctx, client))
	eng := &Engine{cfg: &Config{NotificationManager: notification.NewManager(registry)}}

	first := &message.Message{
		ID:   firstOutgoingMessageID,
		Type: notification.MessageTypeTest,
		Dest: webhook.NewWebhookDest(webURL),
	}
	second := &message.Message{
		ID:   secondOutgoingMessageID,
		Type: notification.MessageTypeTest,
		Dest: webhook.NewWebhookDest(webURL),
	}

	var firstResult *notification.SendResult
	err := retry.DoTemporaryError(func(int) error {
		var err error
		firstResult, err = eng.sendMessage(ctx, first)
		return err
	}, retry.Log(ctx), retry.Limit(2))
	require.NoError(t, err)
	require.NotNil(t, firstResult)
	assert.Equal(t, firstOutgoingMessageID, firstResult.ID)
	assert.Equal(t, notification.StateSent, firstResult.State)

	secondResult, err := eng.sendMessage(ctx, second)
	require.NoError(t, err)
	require.NotNil(t, secondResult)
	assert.Equal(t, secondOutgoingMessageID, secondResult.ID)
	assert.Equal(t, notification.StateSent, secondResult.State)

	require.Len(t, requests, 3)
	assert.Equal(t, firstOutgoingMessageID, requests[0].deliveryID)
	assert.Equal(t, firstOutgoingMessageID, requests[1].deliveryID)
	assert.Equal(t, secondOutgoingMessageID, requests[2].deliveryID)
	assert.NotEqual(t, requests[0].deliveryID, requests[2].deliveryID)
	for _, request := range requests {
		assert.Equal(t, wantBody, request.body)
		assert.NotContains(t, request.body, "Idempotency-Key")
		assert.NotContains(t, request.body, request.deliveryID)
	}

	logs := logOutput.String()
	assert.Contains(t, logs, "webhook request failed")
	assert.NotContains(t, logs, webURL)
	assert.NotContains(t, logs, "opaque-secret-token")
	assert.NotContains(t, logs, "secret-query")
	assert.NotContains(t, logs, wantBody)
	assert.NotContains(t, logs, "sensitive-response-content")
}

const nonWebhookProviderID = "test-non-webhook"

type recordingProvider struct {
	messageIDs []string
}

func (*recordingProvider) ID() string { return nonWebhookProviderID }

func (*recordingProvider) TypeInfo(context.Context) (*nfydest.TypeInfo, error) {
	return &nfydest.TypeInfo{
		Type:    nonWebhookProviderID,
		Name:    "Test non-webhook provider",
		Enabled: true,
	}, nil
}

func (*recordingProvider) ValidateField(context.Context, string, string) error { return nil }

func (*recordingProvider) DisplayInfo(context.Context, map[string]string) (*nfydest.DisplayInfo, error) {
	return &nfydest.DisplayInfo{}, nil
}

func (p *recordingProvider) SendMessage(_ context.Context, msg nfymsg.Message) (*nfymsg.SentMessage, error) {
	p.messageIDs = append(p.messageIDs, msg.MsgID())
	return &nfymsg.SentMessage{State: nfymsg.StateSent}, nil
}

func TestSendMessagePreservesNonWebhookProvider(t *testing.T) {
	const outgoingMessageID = "99999999-8888-4777-8666-555555555555"

	provider := new(recordingProvider)
	registry := nfydest.NewRegistry()
	ctx := (config.Config{}).Context(context.Background())
	registry.RegisterProvider(ctx, provider)
	eng := &Engine{cfg: &Config{NotificationManager: notification.NewManager(registry)}}

	result, err := eng.sendMessage(ctx, &message.Message{
		ID:   outgoingMessageID,
		Type: notification.MessageTypeTest,
		Dest: gadb.NewDestV1(nonWebhookProviderID),
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, outgoingMessageID, result.ID)
	assert.Equal(t, notification.StateSent, result.State)
	assert.Equal(t, []string{outgoingMessageID}, provider.messageIDs)
}
