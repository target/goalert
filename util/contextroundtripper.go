package util

import (
	"context"
	"net/http"
)

type ctxTransport struct {
	rt  http.RoundTripper
	ctx context.Context
}

func (t *ctxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.rt.RoundTrip(req.WithContext(t.ctx))
}

// ContextRoundTripper will return an http.RoundTripper that will replace all request contexts
// with the provided one. This means that values and deadlines for all requests will be bound
// to the original context.
func ContextRoundTripper(ctx context.Context, rt http.RoundTripper) http.RoundTripper {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return &ctxTransport{
		rt:  rt,
		ctx: ctx,
	}
}
