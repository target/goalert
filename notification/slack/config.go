package slack

import (
	"strings"

	"github.com/target/goalert/user"
)

// Config contains values used for the Slack notification sender.
type Config struct {
	BaseURL   string
	UserStore user.Store
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
