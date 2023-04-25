package app

import (
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

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

	http.DefaultTransport = promhttp.InstrumentRoundTripperDuration(promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goalert",
		Subsystem: "http_client",
		Name:      "requests_duration_seconds",
		Help:      "Duration of outgoing HTTP requests in seconds.",
	}, []string{"code", "method"}), http.DefaultTransport)
	http.DefaultTransport = promhttp.InstrumentRoundTripperInFlight(promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goalert",
		Subsystem: "http_client",
		Name:      "requests_in_flight",
		Help:      "Number of outgoing HTTP requests currently active.",
	}), http.DefaultTransport)

	mux.Handle("/metrics", promhttp.Handler())
	srv := http.Server{
		Handler: mux,
	}
	go func() { _ = srv.Serve(l) }()
	return nil
}
