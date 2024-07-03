package slack

import (
	"context"

	"github.com/target/goalert/notification"
)

const FieldSlackUserID = "slack-user-id"

// DMSender wraps a ChannelSender with DM-specific functionality.
type DMSender struct {
	*ChannelSender
}

var _ notification.FriendlyValuer = (*DMSender)(nil)

// DMSender returns a new DMSender wrapping the given ChannelSender.
func (s *ChannelSender) DMSender() *DMSender {
	return &DMSender{s}
}

// FriendlyValue implements notification.FriendlyValuer returning the `@`-handle of the given Slack user ID.
func (s *DMSender) FriendlyValue(ctx context.Context, id string) (string, error) {
	usr, err := s.User(ctx, id)
	if err != nil {
		return "", err
	}

	return "@" + usr.Name, nil
}
