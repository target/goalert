package googlechat

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
)

func testConfig() config.Config {
	var cfg config.Config
	cfg.General.PublicURL = "https://goalert.example"
	return cfg
}

type rewriteTransport struct {
	target *url.URL
	rt     http.RoundTripper
}

func (t rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.target.Scheme
	clone.URL.Host = t.target.Host
	clone.URL.Path = req.URL.Path
	clone.URL.RawPath = req.URL.RawPath
	clone.URL.RawQuery = req.URL.RawQuery
	clone.Host = t.target.Host
	return t.rt.RoundTrip(clone)
}

func TestFormatScheduleOnCallUsers(t *testing.T) {
	ctx := testConfig().Context(context.Background())
	msg := notification.ScheduleOnCallUsers{
		ScheduleName: "Primary Schedule",
		ScheduleURL:  "https://goalert.example/schedules/1",
		Users: []notification.User{
			{ID: "b", Name: "Bravo"},
			{ID: "a", Name: "Alpha"},
		},
	}

	assert.Equal(t,
		"GoAlert on-call shift changed\nSchedule: Primary Schedule\nNow on-call: Alpha, Bravo\nLink: https://goalert.example/schedules/1",
		formatGoAlertMessage(ctx, msg),
	)
}

func TestValidateFieldWebhookURL(t *testing.T) {
	cfg := testConfig()
	ctx := cfg.Context(context.Background())
	sender := &Sender{}

	valid := "https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"
	require.NoError(t, sender.ValidateField(ctx, FieldWebhookURL, valid))

	t.Run("invalid host", func(t *testing.T) {
		err := sender.ValidateField(ctx, FieldWebhookURL, "https://example.com/v1/spaces/AAA/messages?key=k&token=t")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Google Chat incoming webhook URL")
	})

	t.Run("missing token", func(t *testing.T) {
		err := sender.ValidateField(ctx, FieldWebhookURL, "https://chat.googleapis.com/v1/spaces/AAA/messages?key=k")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "key and token")
	})
}

func TestSendMessage(t *testing.T) {
	type result struct {
		state  notification.State
		detail string
	}

	tests := []struct {
		name       string
		message    notification.Message
		wantText   string
		statusCode int
		want       result
	}{
		{
			name: "alert",
			message: notification.Alert{
				Base: nfymsg.Base{
					ID:   "msg-1",
					Dest: NewGoogleChatDest("https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"),
				},
				AlertID:     42,
				Summary:     "Database is down",
				Details:     "postgres is unreachable",
				ServiceName: "Payments",
			},
			wantText: "GoAlert alert\nAlert: #42 Database is down\nService: Payments\nDetails: postgres is unreachable\nLink: https://goalert.example/alerts/42",
			want:     result{state: notification.StateSent},
		},
		{
			name: "alert bundle",
			message: notification.AlertBundle{
				Base: nfymsg.Base{
					ID:   "msg-2",
					Dest: NewGoogleChatDest("https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"),
				},
				ServiceID:   "svc-1",
				ServiceName: "Payments",
				Count:       3,
			},
			wantText: "GoAlert alert bundle\nService: Payments\nCount: 3 unacknowledged alerts\nLink: https://goalert.example/services/svc-1/alerts",
			want:     result{state: notification.StateSent},
		},
		{
			name: "alert status",
			message: notification.AlertStatus{
				Base: nfymsg.Base{
					ID:   "msg-3",
					Dest: NewGoogleChatDest("https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"),
				},
				AlertID:       42,
				Summary:       "Database is down",
				LogEntry:      "acknowledged by on-call",
				NewAlertState: notification.AlertStateAcknowledged,
			},
			wantText: "GoAlert alert update\nAlert: #42 Database is down\nState: acknowledged\nLog: acknowledged by on-call\nLink: https://goalert.example/alerts/42",
			want:     result{state: notification.StateSent},
		},
		{
			name: "on-call",
			message: notification.ScheduleOnCallUsers{
				Base: nfymsg.Base{
					ID:   "msg-4",
					Dest: NewGoogleChatDest("https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"),
				},
				ScheduleName: "Primary Schedule",
				ScheduleURL:  "https://goalert.example/schedules/1",
				Users: []notification.User{
					{ID: "b", Name: "Bravo"},
					{ID: "a", Name: "Alpha"},
				},
			},
			wantText: "GoAlert on-call shift changed\nSchedule: Primary Schedule\nNow on-call: Alpha, Bravo\nLink: https://goalert.example/schedules/1",
			want:     result{state: notification.StateSent},
		},
		{
			name: "temporary failure",
			message: notification.ScheduleOnCallUsers{
				Base: nfymsg.Base{
					ID:   "msg-5",
					Dest: NewGoogleChatDest("https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"),
				},
				ScheduleName: "Primary Schedule",
				ScheduleURL:  "https://goalert.example/schedules/1",
			},
			wantText:   "GoAlert on-call shift changed\nSchedule: Primary Schedule\nNow on-call: Nobody\nLink: https://goalert.example/schedules/1",
			statusCode: http.StatusInternalServerError,
			want:       result{state: notification.StateFailedTemp, detail: "500 Internal Server Error"},
		},
		{
			name: "permanent failure",
			message: notification.ScheduleOnCallUsers{
				Base: nfymsg.Base{
					ID:   "msg-6",
					Dest: NewGoogleChatDest("https://chat.googleapis.com/v1/spaces/AAA/messages?key=k&token=t"),
				},
				ScheduleName: "Primary Schedule",
				ScheduleURL:  "https://goalert.example/schedules/1",
			},
			wantText:   "GoAlert on-call shift changed\nSchedule: Primary Schedule\nNow on-call: Nobody\nLink: https://goalert.example/schedules/1",
			statusCode: http.StatusForbidden,
			want:       result{state: notification.StateFailedPerm, detail: "403 Forbidden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json; charset=UTF-8", r.Header.Get("Content-Type"))

				data, err := io.ReadAll(r.Body)
				require.NoError(t, err)

				var payload ChatMessage
				require.NoError(t, json.NewDecoder(bytes.NewReader(data)).Decode(&payload))
				assert.Equal(t, tt.wantText, payload.Text)

				if tt.statusCode != 0 && tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					_, _ = io.WriteString(w, "chat error")
					return
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			targetURL, err := url.Parse(srv.URL)
			require.NoError(t, err)

			client := &http.Client{
				Transport: rewriteTransport{
					target: targetURL,
					rt:     http.DefaultTransport,
				},
			}

			cfg := testConfig()
			ctx := cfg.Context(context.Background())
			sender := NewSender(ctx, client)

			sent, err := sender.SendMessage(ctx, tt.message)
			require.NoError(t, err)
			require.NotNil(t, sent)
			assert.Equal(t, tt.want.state, sent.State)
			if tt.want.detail == "" {
				assert.Empty(t, sent.StateDetails)
			} else {
				assert.Contains(t, sent.StateDetails, tt.want.detail)
			}
		})
	}
}
