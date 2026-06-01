package graphql2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/config"
)

func TestMapConfigValues_CustomWebhookProviders(t *testing.T) {
	var cfg config.Config
	cfg.GoogleChat.Enable = true
	cfg.CustomWebhook.Enable = true

	values := MapConfigValues(cfg)
	publicValues := MapPublicConfigValues(cfg)

	assertConfigValue := func(t *testing.T, vals []ConfigValue, id string, want string) {
		t.Helper()
		for _, v := range vals {
			if v.ID == id {
				assert.Equal(t, want, v.Value)
				return
			}
		}
		require.Failf(t, "missing config value", "expected %s to be present", id)
	}

	assertConfigValue(t, values, "GoogleChat.Enable", "true")
	assertConfigValue(t, values, "CustomWebhook.Enable", "true")
	assertConfigValue(t, publicValues, "GoogleChat.Enable", "true")
	assertConfigValue(t, publicValues, "CustomWebhook.Enable", "true")
}

func TestApplyConfigValues_CustomWebhookProviders(t *testing.T) {
	cfg, err := ApplyConfigValues(config.Config{}, []ConfigValueInput{
		{ID: "GoogleChat.Enable", Value: "true"},
		{ID: "CustomWebhook.Enable", Value: "true"},
	})
	require.NoError(t, err)
	assert.True(t, cfg.GoogleChat.Enable)
	assert.True(t, cfg.CustomWebhook.Enable)
}
