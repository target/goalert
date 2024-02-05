package alert

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goalert",
		Subsystem: "alert",
		Name:      "created_total",
		Help:      "The total number of created alerts.",
	}, []string{"service_id"})
)
