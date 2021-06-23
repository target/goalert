package remotemonitor

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

type requestIDTransport struct {
	http.RoundTripper
}

func (r *requestIDTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()
	q.Set("x-request-id", uuid.New().String())
	req.URL.RawQuery = q.Encode()

	log.Println(req.Method, req.URL.String())
	return r.RoundTripper.RoundTrip(req)
}
