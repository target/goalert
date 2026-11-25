package slack

import (
	"net/http"

	"github.com/target/goalert/user"
)

// DefaultMaxChannelPages is the default limit for the number of pages loaded when listing Slack channels.
const DefaultMaxChannelPages = 20

// Config contains values used for the Slack notification sender.
type Config struct {
	BaseURL   string
	UserStore *user.Store
	Client    *http.Client

	// MaxChannelPages is the maximum number of pages (200 channels each)
	// to load when listing Slack channels.
	//
	// If zero or negative, DefaultMaxChannelPages is used.
	MaxChannelPages int
}
