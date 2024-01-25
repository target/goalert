package notification

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goalert",
		Subsystem: "notification",
		Name:      "sent_total",
		Help:      "Total number of sent notifications.",
	}, []string{"dest_type", "message_type", "service_id"})
	metricRecvTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "goalert",
		Subsystem: "notification",
		Name:      "recv_total",
		Help:      "Total number of received notification responses.",
	}, []string{"dest_type", "response_type"})
)
