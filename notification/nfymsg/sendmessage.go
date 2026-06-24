package nfymsg

// SentMessage contains information about a message that was sent to a remote
// system.
type SentMessage struct {
	ExternalID   string
	State        State
	StateDetails string
	SrcValue     string
}
