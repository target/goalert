package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/notification/nfymsg"
	"github.com/target/goalert/util/privnet"
)

func TestSender_BlockPrivateAddresses(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
	}))
	defer srv.Close()

	base := http.DefaultTransport.(*http.Transport).Clone()
	base.Proxy = nil
	s := NewSender(context.Background(), &http.Client{Transport: privnet.RoundTripper(base)})
	msg := notification.Test{Base: nfymsg.Base{Dest: NewWebhookDest(srv.URL)}}

	// allowed by default
	var cfg config.Config
	cfg.Webhook.Enable = true
	res, err := s.SendMessage(cfg.Context(context.Background()), msg)
	require.NoError(t, err)
	assert.Equal(t, notification.StateSent, res.State)
	assert.Equal(t, 1, calls)

	// blocked when enabled (test server listens on 127.0.0.1)
	cfg.Webhook.BlockPrivateAddresses = true
	res, err = s.SendMessage(cfg.Context(context.Background()), msg)
	require.NoError(t, err)
	assert.Equal(t, notification.StateFailedPerm, res.State)
	assert.Equal(t, 1, calls, "request should not have been sent")
}
