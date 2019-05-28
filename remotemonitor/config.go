package remotemonitor

// Config contains all necessary values for remote monitoring.
type Config struct {
	// Location is the unique location name of this monitor.
	Location string

	// PublicURL is the publicly-routable base URL for this monitor.
	// It must match what is configured for twilio SMS.
	PublicURL string

	// ListenAddr is the address and port to bind to.
	ListenAddr string

	// CheckMinutes denotes the number of minutes between checks (for all instances).
	CheckMinutes int

	Twilio struct {
		AccountSID string
		AuthToken  string
		FromNumber string
	}

	// Instances determine what remote GoAlert instances will be monitored and send potential errors.
	Instances []Instance
}
