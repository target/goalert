package alert

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/target/goalert/util/timeutil"
)

var (
	metricCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goalert",
		Subsystem: "alert",
		Name:      "created_total",
		Help:      "The total number of created alerts.",
	})
)

type AlertDataPoint struct {
	Timestamp  time.Time
	AlertCount int
}

func SplitRangeByDuration(since, until time.Time, dur timeutil.ISODuration, alerts []Alert) (result []AlertDataPoint) {
	if since.After(until) {
		return result
	}

	iter := since
	for iter.Before(until) {
		dataPoint := AlertDataPoint{Timestamp: iter, AlertCount: 0}
		upperBound := iter.AddDate(dur.Years, dur.Months, dur.Days).Add(dur.TimePart)
		if upperBound.After(until) {
			upperBound = until
		}

		for _, alert := range alerts {
			if !alert.CreatedAt.Before(iter) && alert.CreatedAt.Before(upperBound) {
				dataPoint.AlertCount++
			}
		}
		result = append(result, dataPoint)
		iter = upperBound
	}

	return result
}
