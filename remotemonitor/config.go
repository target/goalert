package remotemonitor

import (
	"github.com/pkg/errors"
	"net/url"
	"strings"
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
	}

	// Instances determine what remote GoAlert instances will be monitored and send potential errors.
	Instances []Instance
}

func (cfg Config) rawCallbackURL(path string, mergeParams ...url.Values) *url.URL {
	base, err := url.Parse(cfg.PublicURL)
	if err != nil {
		panic(errors.Wrap(err, "parse PublicURL"))
	}

	next, err := url.Parse(path)
	if err != nil {
		panic(errors.Wrap(err, "parse path"))
	}

	base.Path = strings.TrimSuffix(base.Path, "/") + "/" + strings.TrimPrefix(next.Path, "/")

	params := base.Query()
	nx := next.Query()
	// set/override any params provided with path
	for name, val := range nx {
		params[name] = val
	}

	// set/override with any additionally provided params
	for _, merge := range mergeParams {
		for name, val := range merge {
			params[name] = val
		}
	}

	base.RawQuery = params.Encode()
	return base
}

// CallbackURL will return a public-routable URL to the given path.
// It will use PublicURL() to fill in missing pieces.
//
// It will panic if provided an invalid URL.
func (cfg Config) CallbackURL(path string, mergeParams ...url.Values) string {
	base := cfg.rawCallbackURL(path, mergeParams...)
	return base.String()
}
