package slack

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
)

func TestValidateRequestSignature(t *testing.T) {
	// Values pulled directly from: https://api.slack.com/authentication/verifying-requests-from-slack
	var cfg config.Config
	cfg.Slack.SigningSecret = "8f742231b10e8888abcd99yyyzzz85a5"

	req, err := http.NewRequestWithContext(cfg.Context(context.Background()), "POST", "http://example.com", strings.NewReader("token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c"))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Slack-Request-Timestamp", "1531420618")
	req.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")

	err = validateRequestSignature(time.Unix(1531420618, 0), req)
	assert.NoError(t, err)

	req, err = http.NewRequestWithContext(cfg.Context(context.Background()), "POST", "http://example.com", strings.NewReader("token=xyzz0WbapA4vBCDEFasx0q6G&team_id=T1DC2JH3J&team_domain=testteamnow&channel_id=G8PSS9T3V&channel_name=foobar&user_id=U2CERLKJA&user_name=roadrunner&command=%2Fwebhook-collect&text=&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FT1DC2JH3J%2F397700885554%2F96rGlfmibIGlgcZRskXaIFfN&trigger_id=398738663015.47445629121.803a0bc887a14d10d2c447fce8b6703c"))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Slack-Request-Timestamp", "15314206189") // changed timestamp
	req.Header.Set("X-Slack-Signature", "v0=a2114d57b48eac39b9ad189dd8316235a7b4a8d21a10bd27519666489c69b503")

	// different timestamp should invalidate the signature
	err = validateRequestSignature(time.Unix(1531420618, 0), req)
	assert.True(t, permission.IsUnauthorized(err), "expected unauthorized error, got: %v", err)
}
