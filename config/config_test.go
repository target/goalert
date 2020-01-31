package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestValidateScopes(t *testing.T) {
	tests := []struct {
		name  string
		value string
		err   string
	}{
		{"empty", "", "does not contain required \"openid\" scope"},
		{"openid missing", "profile email", "does not contain required \"openid\" scope"},
		{"normal", "openid", ""},
		{"multi", "openid profile email", ""},
		{"starts with space", " openid profile email", "starts with extra space"},
		{"ends with space", "openid profile email ", "ends with extra space"},
		{"double spaces", "openid  profile email", "has double spaces"},
		{"double spaces 2", "openid  profile  email", "has double spaces"},
		{"repeating scopes", "openid profile profile", "contains \"profile\" 2 times"},
		{"multiple errors", " openid  profile profile ", "starts with extra space\nends with extra space\nhas double spaces\ncontains \"profile\" 2 times"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateScopes("test", tc.value)
			if tc.err == "" {
				assert.NoError(t, err, "%q", tc.value)
			} else {
				errors := []string{}
				for _, item := range strings.Split(tc.err, "\n") {
					errors = append(errors, "invalid value for 'test': "+item)
				}
				assert.EqualError(t, err, strings.Join(errors, "\n"), "%q", tc.value)
			}
		})
	}
}
