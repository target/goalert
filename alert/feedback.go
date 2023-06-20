package alert

// Feedback represents user provided information about a given alert
type Feedback struct {
	AlertID   int
	Sentiment int
	Note      string
}
