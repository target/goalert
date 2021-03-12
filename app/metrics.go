package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricReqInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "goalert",
		Subsystem: "app",
		Name:      "requests_in_flight",
		Help:      "Current number of requests being served.",
	})
	metricReqTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goalert",
		Subsystem: "app",
		Name:      "requests_total",
		Help:      "Total number of requests by status code.",
	}, []string{"method", "code"})
)
