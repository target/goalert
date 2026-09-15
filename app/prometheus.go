package app

import (
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

type httpClientMetrics struct {
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
}

var httpClientPrometheusMetrics *httpClientMetrics

func initPromServer() error {
	addr := viper.GetString("listen-prometheus")
	if addr == "" {
		return nil
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()

	httpClientPrometheusMetrics = &httpClientMetrics{
		duration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "goalert",
			Subsystem: "http_client",
			Name:      "requests_duration_seconds",
			Help:      "Duration of outgoing HTTP requests in seconds.",
		}, []string{"code", "method"}),
		inFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: "goalert",
			Subsystem: "http_client",
			Name:      "requests_in_flight",
			Help:      "Number of outgoing HTTP requests currently active.",
		}),
	}
	http.DefaultTransport = promhttp.InstrumentRoundTripperDuration(
		httpClientPrometheusMetrics.duration,
		http.DefaultTransport,
	)
	http.DefaultTransport = promhttp.InstrumentRoundTripperInFlight(
		httpClientPrometheusMetrics.inFlight,
		http.DefaultTransport,
	)

	mux.Handle("/metrics", promhttp.Handler())
	srv := http.Server{
		Handler: mux,
	}
	go func() { _ = srv.Serve(l) }()
	return nil
}
