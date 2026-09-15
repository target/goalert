package app

import (
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClientWithWrappedDefaultTransport(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })

	histogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "goalert_test_http_client_requests_duration_seconds",
	}, []string{"code", "method"})

	client, err := newHTTPClient()
	require.NoError(t, err)
	require.NotNil(t, client)
}
