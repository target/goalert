package alert

// Metadata represents user provided information about a given alert
type Metadata struct {
	// ID is the ID of the alert.
	AlertID   int
	Sentiment int
	Note      string
}
