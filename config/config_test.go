package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatchURL(t *testing.T) {
	check := func(valid bool, base, test string) {
		t.Helper()
		result, err := MatchURL(base, test)
		require.Nil(t, err)
		assert.Equalf(t, valid, result, "'%s' should return %t for base URL '%s'", test, valid, base)
	}

	check(true, "http://example.com/", "HTTP://example.COM")
	check(true, "http://example.com", "HTTP://example.COM")
	check(true, "http://example.com", "HTTP://example.COM/child")
	check(true, "http://example.com/", "HTTP://example.COM/child")
	check(true, "http://example.com/", "HTTP://foo:bar@example.COM:80")
	check(true, "http://example.com/?notAllowedQueryParam=&requiredQueryParam=1", "http://example.com?requiredQueryParam=1")

	check(false, "http://example.com/otherchild", "HTTP://example.COM/child")
	check(false, "http://example.com/?notAllowedQueryParam=&requiredQueryParam=1", "http://example.com")
	check(false, "http://example.com/?notAllowedQueryParam=", "http://example.com?notAllowedQueryParam=1")
	check(false, "https://example.com", "http://example.com")
}

func TestValidWebhookURL(t *testing.T) {
	var cfg Config

	check := func(valid bool, url string) {
		t.Helper()
		assert.Equalf(t, valid, cfg.ValidWebhookURL(url), "'%s' should return %t for config %v", url, valid, cfg.Webhook.AllowedURLs)
	}

	// tests when allowedURLs is empty
	check(true, "http://api.example.com")

	cfg.Webhook.AllowedURLs = append(cfg.Webhook.AllowedURLs, "http://api.example.com", "http://subpath.example.com/subpath", "http://reqquery.example.com?req=1")

	// ports must match
	check(false, "http://api.example.com:5555")

	// path must be a subpath match
	check(true, "http://api.example.com/path")
	check(false, "http://subpath.example.com")
	check(false, "http://subpath.example.com/otherpath")
	check(true, "http://subpath.example.com/subpath")
	check(true, "http://subpath.example.com/subpath/2")

	// host must match
	check(false, "http://example.com")

	// scheme must match
	check(false, "https://api.example.com")

	// query must match
	check(true, "http://api.example.com?QueryParam=1")
	check(false, "http://reqquery.example.com")
	check(false, "http://reqquery.example.com?req=2")
	check(true, "http://reqquery.example.com?req=1")

	// implicit ports match (i.e. http :80, https :443)
	check(true, "http://api.example.com:80")
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

func TestConfig_Validate(t *testing.T) {
	assert.NoError(t, Config{}.Validate(), "empty config should always validate")

	t.Run("Twilio.Voice*", func(t *testing.T) {
		var cfg Config
		cfg.Twilio.VoiceName = "Test"
		assert.ErrorContains(t, cfg.Validate(), "Twilio.VoiceLanguage", "language should be required if name is set")

		cfg = Config{}
		cfg.Twilio.VoiceName = "Test"
		cfg.Twilio.VoiceLanguage = "es-US"
		assert.NoError(t, cfg.Validate())

		cfg = Config{}
		cfg.Twilio.VoiceLanguage = "en-US"
		assert.NoError(t, cfg.Validate(), "language alone is valid")

		cfg = Config{}
		cfg.Twilio.VoiceLanguage = "\x00" // non-ASCII value
		assert.Error(t, cfg.Validate(), "language must be a valid string")
	})
}
