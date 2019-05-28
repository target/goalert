package slack

import (
	"strings"
)

// Config contains values used for the Slack notification sender.
type Config struct {
	BaseURL string
}

func (c Config) url(path string) string {
	if c.BaseURL != "" {
		return strings.TrimSuffix(c.BaseURL, "/") + path
	}
	if strings.HasPrefix(path, "/api") {
		return "https://api.slack.com" + path
	}

	return "https://slack.com" + path
}
