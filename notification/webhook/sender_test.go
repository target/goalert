package webhook

import (
	"context"
	"fmt"
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

func TestSender_HTTPStatus(t *testing.T) {
	tests := []struct {
		status int
		state  notification.State
	}{
		{status: http.StatusOK, state: notification.StateSent},
		{status: http.StatusNoContent, state: notification.StateSent},
		{status: http.StatusFound, state: notification.StateFailedPerm},
		{status: http.StatusBadRequest, state: notification.StateFailedPerm},
		{status: http.StatusUnauthorized, state: notification.StateFailedPerm},
		{status: http.StatusNotFound, state: notification.StateFailedPerm},
		{status: http.StatusTooManyRequests, state: notification.StateFailedTemp},
		{status: http.StatusInternalServerError, state: notification.StateFailedTemp},
		{status: http.StatusServiceUnavailable, state: notification.StateFailedTemp},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("status_%d", test.status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
			}))
			defer srv.Close()

			var cfg config.Config
			cfg.Webhook.Enable = true
			s := NewSender(context.Background(), srv.Client())
			msg := notification.Test{Base: nfymsg.Base{Dest: NewWebhookDest(srv.URL)}}

			res, err := s.SendMessage(cfg.Context(context.Background()), msg)
			require.NoError(t, err)
			require.Equal(t, test.state, res.State)
			if test.state.IsOK() {
				assert.Empty(t, res.StateDetails)
			} else {
				assert.Contains(t, res.StateDetails, fmt.Sprint(test.status))
			}
		})
	}
}

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
