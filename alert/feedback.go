package alert

// Feedback represents user provided information about a given alert
type Feedback struct {
	// ID is the ID of the alert.
	AlertID   int
	Sentiment int
	Note      string
}
