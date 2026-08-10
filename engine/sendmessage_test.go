package engine

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/alert/alertlog"
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
		assert.NotContains(t, request.body, "AlertState")
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

type alertStatusTestDB struct {
	logType alertlog.Type
}

type alertStatusTestDriver struct {
	source *alertStatusTestDB
}

func (d alertStatusTestDriver) Open(string) (driver.Conn, error) {
	return &alertStatusTestConn{source: d.source}, nil
}

type alertStatusTestConnector struct {
	source *alertStatusTestDB
}

func (c alertStatusTestConnector) Connect(context.Context) (driver.Conn, error) {
	return &alertStatusTestConn{source: c.source}, nil
}

func (c alertStatusTestConnector) Driver() driver.Driver {
	return alertStatusTestDriver(c)
}

type alertStatusTestConn struct {
	source *alertStatusTestDB
}

func (c *alertStatusTestConn) Prepare(query string) (driver.Stmt, error) {
	return &alertStatusTestStmt{source: c.source, query: query}, nil
}

func (*alertStatusTestConn) Close() error { return nil }

func (*alertStatusTestConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported by the alert status test database")
}

type alertStatusTestStmt struct {
	source *alertStatusTestDB
	query  string
}

func (*alertStatusTestStmt) Close() error { return nil }

func (*alertStatusTestStmt) NumInput() int { return -1 }

func (*alertStatusTestStmt) Exec([]driver.Value) (driver.Result, error) {
	return nil, errors.New("exec is not supported by the alert status test database")
}

func (s *alertStatusTestStmt) Query([]driver.Value) (driver.Rows, error) {
	return s.source.rows(s.query)
}

type alertStatusTestRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *alertStatusTestRows) Columns() []string { return r.columns }

func (*alertStatusTestRows) Close() error { return nil }

func (r *alertStatusTestRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func (db *alertStatusTestDB) rows(query string) (driver.Rows, error) {
	now := time.Date(2026, time.August, 10, 8, 0, 0, 0, time.UTC)
	var values []driver.Value
	switch {
	case strings.Contains(strings.ToLower(query), "from alert_logs log"):
		values = []driver.Value{
			int64(101), int64(42), now, string(db.logType), "", nil, nil, nil,
			nil, nil, nil, nil, nil, nil, "", nil,
		}
	case strings.Contains(strings.ToLower(query), "from alerts a"):
		values = []driver.Value{
			int64(42), "Synthetic summary", "Synthetic details",
			"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", "manual", "active", now, nil,
		}
	case strings.Contains(strings.ToLower(query), "-- name: nfyoriginalmessagestatus"):
		values = []driver.Value{
			int64(42), nil, nil, nil, now, nil, nil, nil,
			"22222222-3333-4444-8555-666666666666", "sent", now, "alert_notification",
			nil, nil, int64(0), int64(0), nil, nil, now, nil, nil, nil, "", nil, nil,
			nil, nil,
		}
	default:
		return nil, errors.New("unexpected query in alert status test database")
	}

	return &alertStatusTestRows{
		columns: make([]string, len(values)),
		values:  [][]driver.Value{values},
	}, nil
}

func TestSendMessagePropagatesAlertStateToWebhookRequest(t *testing.T) {
	const (
		outgoingMessageID = "11111111-2222-4333-8444-555555555555"
		webURL            = "https://gateway.invalid/v1/goalert/contact-method/opaque-secret-token?route=secret-query"
	)
	tests := []struct {
		name      string
		logType   alertlog.Type
		wireState string
	}{
		{name: "acknowledged", logType: alertlog.TypeAcknowledged, wireState: "Acknowledged"},
		{name: "escalated", logType: alertlog.TypeEscalated, wireState: "Unacknowledged"},
		{name: "closed", logType: alertlog.TypeClosed, wireState: "Closed"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var appConfig config.Config
			appConfig.Webhook.Enable = true
			ctx := appConfig.Context(context.Background())
			registry := nfydest.NewRegistry()
			var requestCount int
			client := &http.Client{
				Transport: deliveryRoundTripFunc(func(req *http.Request) (*http.Response, error) {
					requestCount++
					assert.Equal(t, outgoingMessageID, req.Header.Get("Idempotency-Key"))
					body, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					var payload map[string]interface{}
					require.NoError(t, json.Unmarshal(body, &payload))
					assert.Equal(t, test.wireState, payload["AlertState"])
					assert.NotContains(t, payload, "NewAlertState")
					return &http.Response{
						StatusCode: http.StatusAccepted,
						Header:     make(http.Header),
						Body:       io.NopCloser(strings.NewReader("ignored")),
						Request:    req,
					}, nil
				}),
			}
			registry.RegisterProvider(ctx, webhook.NewSender(ctx, client))

			db := sql.OpenDB(alertStatusTestConnector{source: &alertStatusTestDB{logType: test.logType}})
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			alertLogStore, err := alertlog.NewStore(ctx, db, registry)
			require.NoError(t, err)
			alertStore, err := alert.NewStore(ctx, db, alertLogStore, nil)
			require.NoError(t, err)
			notificationStore, err := notification.NewStore(ctx, db)
			require.NoError(t, err)

			eng := &Engine{cfg: &Config{
				AlertLogStore:       alertLogStore,
				AlertStore:          alertStore,
				NotificationStore:   notificationStore,
				NotificationManager: notification.NewManager(registry),
			}}
			result, err := eng.sendMessage(ctx, &message.Message{
				ID:         outgoingMessageID,
				Type:       notification.MessageTypeAlertStatus,
				Dest:       webhook.NewWebhookDest(webURL),
				AlertID:    42,
				AlertLogID: 101,
			})
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, outgoingMessageID, result.ID)
			assert.Equal(t, notification.StateSent, result.State)
			assert.Equal(t, 1, requestCount)
		})
	}
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
