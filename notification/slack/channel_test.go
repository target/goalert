package slack

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
)

func TestChannelSender_LoadChannels(t *testing.T) {
	var waitUntil time.Time
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users.conversations", func(w http.ResponseWriter, r *http.Request) {
		switch r.FormValue("cursor") {
		case "":
			_, _ = io.WriteString(w, `{"ok":true,"channels":[{"id":"C1","name":"channel1"},{"id":"C2","name":"channel2"}],"response_metadata":{"next_cursor":"cursor_1"}}`)
		case "cursor_1":
			_, _ = io.WriteString(w, `{"ok":true,"channels":[{"id":"C3","name":"channel3"},{"id":"C4","name":"channel4"}],"response_metadata":{"next_cursor":"cursor_2"}}`)
		case "cursor_2":
			// ensure retry/delay logic works
			if waitUntil.IsZero() {
				waitUntil = time.Now().Add(time.Second)
				w.Header().Set("Retry-After", "1")
				w.WriteHeader(429)
				return
			}
			if time.Until(waitUntil) > 0 {
				t.Error("failed to respect Retry-After value")
			}
			_, _ = io.WriteString(w, `{"ok":true,"channels":[{"id":"C5","name":"channel5"}]}`)
		default:
			t.Errorf("unexpected cursor value '%s'", r.FormValue("cursor"))
		}
	})
	mux.HandleFunc("/api/auth.test", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"ok":true,"team_id":"team_1"}`)
	})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Log("Request:", r.URL.String())
		mux.ServeHTTP(w, r)
	}))
	defer srv.Close()

	var cfg config.Config
	cfg.Slack.AccessToken = "access_token"
	ctx := cfg.Context(context.Background())

	sender, err := NewChannelSender(ctx, Config{BaseURL: srv.URL, Client: http.DefaultClient})
	require.NoError(t, err)

	ch, err := sender.loadChannels(ctx)
	require.NoError(t, err)

	assert.ElementsMatch(t, []Channel{
		{ID: "C1", Name: "#channel1", TeamID: "team_1"},
		{ID: "C2", Name: "#channel2", TeamID: "team_1"},
		{ID: "C3", Name: "#channel3", TeamID: "team_1"},
		{ID: "C4", Name: "#channel4", TeamID: "team_1"},
		{ID: "C5", Name: "#channel5", TeamID: "team_1"},
	}, ch)
}
