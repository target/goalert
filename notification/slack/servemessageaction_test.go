package slack

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
)

func TestValidateRequestSignature(t *testing.T) {
	var cfg config.Config
	cfg.Slack.SigningSecret = "secret"

	ts := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	v := make(url.Values)
	v.Set("foo", "bar")

	req, err := http.NewRequestWithContext(cfg.Context(context.Background()), "POST", "http://example.com", strings.NewReader(v.Encode()))
	require.NoError(t, err)

	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(ts.Unix(), 10))
	req.Header.Set("X-Slack-Signature", "v0=890648c7d227bf188d0261b6ab30587b0dbd39a3a1e24c83e38050aa6a3b6bb9")

	err = validateRequestSignature(ts, req)
	assert.NoError(t, err)

	// different timestamp should invalidate the signature
	req.Header.Set("X-Slack-Request-Timestamp", strconv.FormatInt(ts.Add(time.Second).Unix(), 10))
	err = validateRequestSignature(ts, req)
	assert.Error(t, err)
}
