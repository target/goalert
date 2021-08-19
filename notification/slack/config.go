package slack

import (
	"strings"

	"github.com/target/goalert/auth"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/user"
)

// Config contains values used for the Slack notification sender and handler.
type Config struct {
	BaseURL     string
	UserStore   user.Store // ok to have
	AuthHandler auth.Handler
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

// ReceiverSetter is an optional interface a Sender can implement for use with two-way interactions.
type ReceiverSetter interface {
	SetReceiver(notification.Receiver)
}
