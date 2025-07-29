package calllimiter

import (
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
)

type roundTripper struct {
	rt http.RoundTripper
}

func RoundTripper(rt http.RoundTripper) http.RoundTripper {
	return &roundTripper{rt: rt}
}

type ErrCallLimitReached struct {
	NumCalls int
}

func (e *ErrCallLimitReached) ClientError() bool { return true }

func (e *ErrCallLimitReached) Error() string {
	return fmt.Sprintf("external call limit reached (%d calls)", e.NumCalls)
}

var ErrHTTPCallLimitReached = errors.New("HTTP call limit reached")

func (r *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	l := FromContext(req.Context())
	if !l.Allow() {
		return nil, &ErrCallLimitReached{NumCalls: int(atomic.LoadInt64(&l.numCalls))}
	}
	return r.rt.RoundTrip(req)
}
