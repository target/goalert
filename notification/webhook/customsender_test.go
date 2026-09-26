package webhook

import (
	"context"
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

func TestCustomSender_SendMessage(t *testing.T) {
	type result struct {
		state  notification.State
		detail string
	}

	tests := []struct {
		name       string
		statusCode int
		want       result
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			want:       result{state: notification.StateSent},
		},
		{
			name:       "temporary failure",
			statusCode: http.StatusInternalServerError,
			want:       result{state: notification.StateFailedTemp, detail: "500 Internal Server Error"},
		},
		{
			name:       "permanent failure",
			statusCode: http.StatusForbidden,
			want:       result{state: notification.StateFailedPerm, detail: "403 Forbidden"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Contains(t, string(body), "Database is down")
				assert.Contains(t, string(body), "Payments")

				if tt.statusCode != http.StatusOK {
					w.WriteHeader(tt.statusCode)
					_, _ = io.WriteString(w, "custom webhook error")
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer srv.Close()

			cfg := config.Config{}
			cfg.General.PublicURL = "https://goalert.example"
			ctx := cfg.Context(context.Background())

			targetURL, err := url.Parse(srv.URL)
			require.NoError(t, err)

			sender := &CustomSender{Client: srv.Client()}
			msg := notification.Alert{
				Base: nfymsg.Base{
					ID:   "msg-1",
					Dest: NewCustomWebhookDest(targetURL.String(), `{"text":"{{.Summary}} - {{.ServiceName}}"}`, "application/json"),
				},
				AlertID:     42,
				Summary:     "Database is down",
				Details:     "postgres is unreachable",
				ServiceName: "Payments",
			}

			sent, err := sender.SendMessage(ctx, msg)
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

func TestCustomSender_ValidateField(t *testing.T) {
	cfg := config.Config{}
	cfg.General.PublicURL = "https://goalert.example"
	ctx := cfg.Context(context.Background())
	sender := &CustomSender{}

	require.NoError(t, sender.ValidateField(ctx, FieldWebhookURL, "https://example.com"))
	require.NoError(t, sender.ValidateField(ctx, FieldBodyTemplate, `{"text":"{{.Summary}}"}`))
	require.NoError(t, sender.ValidateField(ctx, FieldContentType, "application/json"))
	require.NoError(t, sender.ValidateField(ctx, FieldContentType, ""))
}
