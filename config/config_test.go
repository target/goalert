package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchURL(t *testing.T) {
	val, err := MatchURL("http://example.com/", "HTTP://example.COM")
	assert.True(t, val)
	assert.Nil(t, err)

	val, err = MatchURL("http://example.com/", "HTTP://foo:bar@example.COM:80")
	assert.True(t, val)
	assert.Nil(t, err)

	val, err = MatchURL("http://example.com/?notAllowedQueryParam=&requiredQueryParam=1", "http://example.com?requiredQueryParam=1")
	assert.True(t, val)
	assert.Nil(t, err)

	val, err = MatchURL("http://example.com/?notAllowedQueryParam=&requiredQueryParam=1", "http://example.com")
	assert.False(t, val)
	assert.Error(t, err)

	val, err = MatchURL("http://example.com/?notAllowedQueryParam=", "http://example.com?notAllowedQueryParam=1")
	assert.False(t, val)
	assert.Error(t, err)

	val, err = MatchURL("https://example.com", "http://example.com")
	assert.False(t, val)
	assert.Error(t, err)

}

func TestValidWebhookURL(t *testing.T) {
	var cfg Config
	// tests when allowedURLs is empty
	assert.True(t, cfg.ValidWebhookURL("http://api.example.com"))

	cfg.Webhook.AllowedURLs = append(cfg.Webhook.AllowedURLs, "http://api.example.com")
	// tests when allowedURLs has been set
	assert.True(t, cfg.ValidWebhookURL("http://api.example.com:5555/path"))
	assert.False(t, cfg.ValidWebhookURL("http://example.com"))
}

func TestValidReferer(t *testing.T) {
	t.Run("default config same host", func(t *testing.T) {
		var cfg Config
		cfg.General.PublicURL = "https://example.com"

		assert.True(t, cfg.ValidReferer("http://foobar.com/bin", "http://foobar.com/baz"), "same request host")
		assert.False(t, cfg.ValidReferer("https://foobar.com/bin", "http://foobar.com/baz"), "same request host, diff schema")
		assert.True(t, cfg.ValidReferer("https://foobar.com/bin", "https://example.com/baz"), "valid public URL match")
		assert.False(t, cfg.ValidReferer("https://foobar.com/bin", "http://example.com/baz"), "public URL match, wrong schema")
	})

	t.Run("auth referers set", func(t *testing.T) {
		var cfg Config
		cfg.General.PublicURL = "https://example.com"
		cfg.Auth.RefererURLs = []string{
			"http://foobar.com",
			"https://binbaz.com",
			"http://path.com/foo",
		}

		assert.True(t, cfg.ValidReferer("https://req.com", "https://example.com/foo"), "pub URL match")
		assert.True(t, cfg.ValidReferer("https://req.com", "http://foobar.com"), "auth URL match")
		assert.True(t, cfg.ValidReferer("https://req.com", "https://binbaz.com/foo"), "auth URL prefix match")
		assert.True(t, cfg.ValidReferer("https://req.com", "http://path.com/foo/bar"), "auth URL path match")

		assert.False(t, cfg.ValidReferer("https://req.com", "https://foobar.com"), "auth schema mismatch")
		assert.False(t, cfg.ValidReferer("https://req.com", "http://path.com/bar"), "auth URL path mismatch")
		assert.False(t, cfg.ValidReferer("https://req.com", "https://req.com/bar"), "auth URL set (no same host)")
	})
}
