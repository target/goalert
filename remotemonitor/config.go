package remotemonitor

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

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
		MessageSID string
	}

	SMTP struct {
		From string

		// ServerAddr is the address of the SMTP server, including port.
		ServerAddr string
		User, Pass string

		// Retries is the number of times to retry sending with an email integration key.
		Retries int
	}

	// Instances determine what remote GoAlert instances will be monitored and send potential errors.
	Instances []Instance
}

func (cfg Config) Validate() error {
	if cfg.Location == "" {
		return errors.New("location is required")
	}
	if cfg.PublicURL == "" {
		return errors.New("public URL is required")
	}
	_, err := url.Parse(cfg.PublicURL)
	if err != nil {
		return fmt.Errorf("parse public URL: %v", err)
	}

	if cfg.ListenAddr == "" {
		return errors.New("listen address is required")
	}

	if cfg.CheckMinutes < 1 {
		return errors.New("check minutes is required")
	}

	if cfg.Twilio.AccountSID == "" {
		return errors.New("twilio account SID is required")
	}
	if cfg.Twilio.AuthToken == "" {
		return errors.New("twilio auth token is required")
	}
	if cfg.Twilio.FromNumber == "" {
		return errors.New("twilio from number is required")
	}

	var hasEmail bool
	for idx, i := range cfg.Instances {
		if err := i.Validate(); err != nil {
			return fmt.Errorf("instance[%d] %q: %v", idx, i.Location, err)
		}
		if i.EmailAPIKey != "" {
			hasEmail = true
		}
	}

	if !hasEmail {
		return nil
	}

	if cfg.SMTP.ServerAddr == "" {
		return errors.New("SMTP server address is required")
	}
	if _, _, err := net.SplitHostPort(cfg.SMTP.ServerAddr); err != nil {
		return fmt.Errorf("parse SMTP server address: %v", err)
	}
	if cfg.SMTP.From == "" {
		return errors.New("SMTP from address is required")
	}
	if cfg.SMTP.Retries > 5 || cfg.SMTP.Retries < 0 {
		return errors.New("SMTP retries must be between 0 and 5")
	}

	return nil
}
