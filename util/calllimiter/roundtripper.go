package calllimiter

import (
	"errors"
	"net/http"
)

type roundTripper struct {
	rt http.RoundTripper
}

func RoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &roundTripper{rt: rt}
}

var ErrHTTPCallLimitReached = errors.New("HTTP call limit reached")

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !FromContext(req.Context()).Allow() {
		return nil, ErrHTTPCallLimitReached
	}
	return r.rt.RoundTrip(req)
}
