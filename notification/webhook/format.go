package webhook

import (
	"context"
	"net/url"
	"strings"

	"github.com/target/goalert/notification"
)

var _ notification.FriendlyValuer = Sender{}

// MaskURLPass will mask the password (if any) in the URL>
func MaskURLPass(u *url.URL) string {
	if u.User == nil {
		return u.String()
	}

	_, ok := u.User.Password()
	if !ok {
		return u.String()
	}

	u.User = url.UserPassword(u.User.Username(), "")

	parts := strings.SplitN(u.String(), "@", 2)
	parts[0] += "***"
	return strings.Join(parts, "@")
}

// FriendlyValue will return a display-ready version of the URL with the password masked,
// query removed, and the path truncated to 15 chars.
func (Sender) FriendlyValue(ctx context.Context, value string) (string, error) {
	u, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	if len(u.Path) > 15 {
		u.Path = u.Path[:12] + "..."
	}

	return MaskURLPass(u), nil
}
