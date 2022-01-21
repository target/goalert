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

// SplitRangeByDuration requires that alerts is sorted by CreatedAt
func SplitRangeByDuration(since, until time.Time, dur timeutil.ISODuration, alerts []Alert) (result []AlertDataPoint) {
	if since.After(until) {
		return result
	}

	i := 0
	// ignore alerts created before since
	for i < len(alerts) && alerts[i].CreatedAt.Before(since) {
		i++
	}

	ts, upperBound := since, since.AddDate(dur.Years, dur.Months, dur.Days).Add(dur.TimePart)
	if upperBound.After(until) {
		upperBound = until
	}

	for ts.Before(until) {
		next := AlertDataPoint{Timestamp: ts, AlertCount: 0}
		for i < len(alerts) && alerts[i].CreatedAt.Before(upperBound) {
			next.AlertCount++
			i++
		}
		result = append(result, next)
		ts, upperBound = upperBound, upperBound.AddDate(dur.Years, dur.Months, dur.Days).Add(dur.TimePart)
		if upperBound.After(until) {
			upperBound = until
		}
	}

	return result
}
