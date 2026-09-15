package app

import (
	"errors"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/target/goalert/util/calllimiter"
	"github.com/target/goalert/util/privnet"
)

// initialDefaultTransport keeps the standard transport available after another
// initializer wraps http.DefaultTransport (for example, Prometheus metrics).
var initialDefaultTransport = http.DefaultTransport

func cloneDefaultHTTPTransport() (*http.Transport, error) {
	if transport, ok := http.DefaultTransport.(*http.Transport); ok {
		return transport.Clone(), nil
	}
	if transport, ok := initialDefaultTransport.(*http.Transport); ok {
		return transport.Clone(), nil
	}

	return nil, errors.New("http.DefaultTransport is not an *http.Transport")
}

func newHTTPClient() (*http.Client, error) {
	base, err := cloneDefaultHTTPTransport()
	if err != nil {
		return nil, err
	}

	transport := privnet.RoundTripper(base)
	if metrics := httpClientPrometheusMetrics; metrics != nil {
		transport = promhttp.InstrumentRoundTripperDuration(metrics.duration, transport)
		transport = promhttp.InstrumentRoundTripperInFlight(metrics.inFlight, transport)
	}
	transport = calllimiter.RoundTripper(transport)

	return &http.Client{Transport: transport}, nil
}
