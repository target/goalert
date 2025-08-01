package slack

import (
	"net/http"

	"github.com/target/goalert/user"
)

// Config contains values used for the Slack notification sender.
type Config struct {
	BaseURL   string
	UserStore *user.Store
	Client    *http.Client
}
