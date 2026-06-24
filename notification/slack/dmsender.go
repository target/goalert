package slack

// DMSender wraps a ChannelSender with DM-specific functionality.
type DMSender struct {
	*ChannelSender
}

// DMSender returns a new DMSender wrapping the given ChannelSender.
func (s *ChannelSender) DMSender() *DMSender {
	return &DMSender{s}
}
