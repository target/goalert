package engine

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricCycleTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goalert",
		Subsystem: "engine",
		Name:      "cycle_total",
		Help:      "Total number of engine cycles.",
	})

	metricModuleDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goalert",
		Subsystem: "engine",
		Name:      "cycle_duration_seconds",
		Help:      "Engine cycle duration in seconds by module.",
	}, []string{"module"})
)
