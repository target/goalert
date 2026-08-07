package ntfy

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfymsg"
)

// testConfig points at the given ntfy server and gives CallbackURL something to build on.
func testConfig(serverURL string) config.Config {
	var cfg config.Config
	cfg.General.PublicURL = "https://goalert.example.com"
	cfg.Ntfy.Enable = true
	cfg.Ntfy.ServerURL = serverURL
	cfg.Ntfy.Token = "tok-123"

	return cfg
}

func alertMsg(meta map[string]string) nfymsg.Alert {
	return nfymsg.Alert{
		Base:        nfymsg.Base{ID: "msg-1", Dest: NewNtfyDest("alerts-jane")},
		AlertID:     42,
		Summary:     "Disk full",
		Details:     "Root volume is at 98%.",
		ServiceID:   "svc-1",
		ServiceName: "demo",
		Meta:        meta,
	}
}

func TestSendMessage_Alert(t *testing.T) {
	var gotPath, gotTitle, gotPriority, gotClick, gotAuth, gotBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotPath, gotBody = r.URL.Path, string(body)
		gotTitle = r.Header.Get("X-Title")
		gotPriority = r.Header.Get("X-Priority")
		gotClick = r.Header.Get("X-Click")
		gotAuth = r.Header.Get("Authorization")
	}))
	defer srv.Close()

	cfg := testConfig(srv.URL)
	s := NewSender(context.Background(), srv.Client())

	res, err := s.SendMessage(cfg.Context(context.Background()),
		alertMsg(map[string]string{MetaClickURL: "https://chat.example.com/demo/pl/abc123"}))
	require.NoError(t, err)
	assert.Equal(t, nfymsg.StateSent, res.State)

	assert.Equal(t, "/alerts-jane", gotPath)
	assert.Equal(t, "Disk full", gotTitle)
	assert.Equal(t, "Root volume is at 98%.", gotBody)
	// Alerts page at max priority so they break through Do Not Disturb.
	assert.Equal(t, "5", gotPriority)
	assert.Equal(t, "https://chat.example.com/demo/pl/abc123", gotClick)
	assert.Equal(t, "Bearer tok-123", gotAuth)
}

func TestSendMessage_ClickFallsBackToGoAlert(t *testing.T) {
	for _, tc := range []struct {
		name string
		meta map[string]string
	}{
		{"no meta", nil},
		{"empty value", map[string]string{MetaClickURL: ""}},
		{"not absolute", map[string]string{MetaClickURL: "/demo/pl/abc123"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotClick string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotClick = r.Header.Get("X-Click")
			}))
			defer srv.Close()

			cfg := testConfig(srv.URL)
			s := NewSender(context.Background(), srv.Client())

			_, err := s.SendMessage(cfg.Context(context.Background()), alertMsg(tc.meta))
			require.NoError(t, err)
			assert.Equal(t, "https://goalert.example.com/alerts/42", gotClick)
		})
	}
}

// A summary spanning lines would otherwise make the request unsendable, since the title is a header.
func TestSendMessage_MultilineSummary(t *testing.T) {
	var gotTitle string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTitle = r.Header.Get("X-Title")
	}))
	defer srv.Close()

	cfg := testConfig(srv.URL)
	s := NewSender(context.Background(), srv.Client())

	msg := alertMsg(nil)
	msg.Summary = "Disk full\non node-1\r\nagain"

	res, err := s.SendMessage(cfg.Context(context.Background()), msg)
	require.NoError(t, err)
	assert.Equal(t, nfymsg.StateSent, res.State)
	assert.Equal(t, "Disk full on node-1  again", gotTitle)
}

func TestSendMessage_Verification(t *testing.T) {
	var gotBody, gotPriority string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		gotPriority = r.Header.Get("X-Priority")
	}))
	defer srv.Close()

	cfg := testConfig(srv.URL)
	s := NewSender(context.Background(), srv.Client())

	_, err := s.SendMessage(cfg.Context(context.Background()), nfymsg.Verification{
		Base: nfymsg.Base{ID: "msg-1", Dest: NewNtfyDest("alerts-jane")},
		Code: "123456",
	})
	require.NoError(t, err)
	assert.Contains(t, gotBody, "123456")
	assert.Equal(t, "4", gotPriority)
}

func TestSendMessage_RejectsPermanentlyOn4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	cfg := testConfig(srv.URL)
	s := NewSender(context.Background(), srv.Client())

	res, err := s.SendMessage(cfg.Context(context.Background()), alertMsg(nil))
	require.NoError(t, err)
	assert.Equal(t, nfymsg.StateFailedPerm, res.State)
}

// A 5xx may succeed on a retry, so it must surface as an error rather than a permanent failure.
func TestSendMessage_RetriesOn5xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	cfg := testConfig(srv.URL)
	s := NewSender(context.Background(), srv.Client())

	_, err := s.SendMessage(cfg.Context(context.Background()), alertMsg(nil))
	assert.Error(t, err)
}

func TestSendMessage_NoServerURLConfigured(t *testing.T) {
	cfg := testConfig("")
	s := NewSender(context.Background(), http.DefaultClient)

	res, err := s.SendMessage(cfg.Context(context.Background()), alertMsg(nil))
	require.NoError(t, err)
	assert.Equal(t, nfymsg.StateFailedPerm, res.State)
}

func TestValidateTopic(t *testing.T) {
	for _, tc := range []struct {
		topic string
		ok    bool
	}{
		{"alerts-jane", true},
		{"a_b-C9", true},
		{"", false},
		{"has space", false},
		{"has/slash", false},
		{"dots.not.allowed", false},
		{"ünïcode", false},
	} {
		t.Run(tc.topic, func(t *testing.T) {
			err := validateTopic(tc.topic)
			if tc.ok {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestTopicURL(t *testing.T) {
	assert.Equal(t, "https://ntfy.example.com/alerts-jane",
		topicURL("https://ntfy.example.com", "alerts-jane"))
	assert.Equal(t, "https://ntfy.example.com/alerts-jane",
		topicURL("https://ntfy.example.com/", "alerts-jane"))
	assert.Empty(t, topicURL("", "alerts-jane"))
	assert.Empty(t, topicURL("https://ntfy.example.com", ""))
}
