package graphqlapp

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	metricResolverHist = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "goalert",
		Subsystem: "graphql",
		Name:      "resolver_",
		Help:      "GraphQL resolver statistics.",
	}, []string{"name", "error"})
)
