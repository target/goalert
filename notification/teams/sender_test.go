package teams

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
)

func TestSender_SendMessage(t *testing.T) {
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		data, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		require.NoError(t, json.Unmarshal(data, &body))
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	ctx := testCtx()
	s := NewSender(ctx, srv.Client())

	res, err := s.SendMessage(ctx, notification.Alert{
		Base: nfymsg.Base{
			ID:   "msg-id",
			Dest: NewTeamsWorkflowDest(srv.URL),
		},
		AlertID:     123,
		Summary:     "Example Summary",
		ServiceName: "Example Service",
	})
	require.NoError(t, err)
	assert.Equal(t, notification.StateSent, res.State)

	assert.Equal(t, "message", body["type"])
	atts, ok := body["attachments"].([]any)
	require.True(t, ok)
	require.Len(t, atts, 1)
	att, ok := atts[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "application/vnd.microsoft.card.adaptive", att["contentType"])
	card, ok := att["content"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "AdaptiveCard", card["type"])
}

func TestSender_SendMessage_PermanentFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	ctx := testCtx()
	s := NewSender(ctx, srv.Client())

	res, err := s.SendMessage(ctx, notification.Test{
		Base: nfymsg.Base{Dest: NewTeamsWorkflowDest(srv.URL)},
	})
	require.NoError(t, err)
	assert.Equal(t, notification.StateFailedPerm, res.State)
}

func TestSender_SendMessage_TemporaryFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	ctx := testCtx()
	s := NewSender(ctx, srv.Client())

	_, err := s.SendMessage(ctx, notification.Test{
		Base: nfymsg.Base{Dest: NewTeamsWorkflowDest(srv.URL)},
	})
	require.Error(t, err)
}

func TestSender_SendMessage_DisallowedURL(t *testing.T) {
	var cfg config.Config
	cfg.General.PublicURL = "http://example.com"
	cfg.Teams.Enable = true
	cfg.Teams.AllowedWorkflowURLs = []string{"https://allowed.example.com"}
	ctx := cfg.Context(context.Background())

	s := NewSender(ctx, nil)

	res, err := s.SendMessage(ctx, notification.Test{
		Base: nfymsg.Base{Dest: NewTeamsWorkflowDest("https://evil.example.com/hook")},
	})
	require.NoError(t, err)
	assert.Equal(t, notification.StateFailedPerm, res.State)
}

func TestSender_ValidateField(t *testing.T) {
	ctx := testCtx()
	var s Sender

	assert.NoError(t, s.ValidateField(ctx, FieldWebhookURL, "https://example.westus.logic.azure.com/workflows/abc"))
	assert.Error(t, s.ValidateField(ctx, FieldWebhookURL, "http://example.com/insecure"), "http URLs should be rejected")
	assert.Error(t, s.ValidateField(ctx, FieldWebhookURL, "not-a-url"))
	assert.Error(t, s.ValidateField(ctx, "unknown_field", "value"))

	var cfg config.Config
	cfg.Teams.AllowedWorkflowURLs = []string{"https://allowed.example.com"}
	restrictedCtx := cfg.Context(context.Background())
	assert.NoError(t, s.ValidateField(restrictedCtx, FieldWebhookURL, "https://allowed.example.com/hook"))
	assert.Error(t, s.ValidateField(restrictedCtx, FieldWebhookURL, "https://other.example.com/hook"))
}

func TestSender_DisplayInfo(t *testing.T) {
	var s Sender
	info, err := s.DisplayInfo(testCtx(), map[string]string{
		FieldWebhookURL: "https://example.westus.logic.azure.com/workflows/abc",
	})
	require.NoError(t, err)
	assert.Equal(t, "example.westus.logic.azure.com", info.Text)
	assert.Equal(t, FallbackIconURL, info.IconURL)
}
