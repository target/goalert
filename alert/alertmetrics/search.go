package alertmetrics

import (
	"context"
	"database/sql"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions contains criteria for filtering and sorting alert metrics.
type SearchOptions struct {
	// ServiceIDs, if specified, will restrict alert metrics to those with a matching ServiceID on IDs, if valid.
	ServiceIDs []string `json:"v,omitempty"`

	// Since will omit any alert metrics created any time before the provided time.
	Since time.Time `json:"n,omitempty"`

	// Until will only include alert metrics that were created before the provided time.
	Until time.Time `json:"b,omitempty"`
}

var searchTemplate = template.Must(template.New("alert-metrics-search").Funcs(search.Helpers()).Parse(`
	SELECT
		service_id,
		closed_at::date,
		count(*),
		EXTRACT(EPOCH FROM coalesce(avg(time_to_ack), avg(time_to_close))),
		EXTRACT(EPOCH FROM avg(time_to_close))
	FROM alert_metrics
	WHERE true
	{{if .ServiceIDs}}
		AND service_id = any(:services)
	{{end}}
	{{ if not .Until.IsZero }}
		AND closed_at::date < :until
	{{ end }}
	{{ if not .Since.IsZero }}
		AND closed_at::date >= :since
	{{ end }}
	GROUP BY service_id, closed_at::date
	ORDER BY closed_at
`))

type renderData SearchOptions

func (opts renderData) Normalize() (*renderData, error) {
	err := validate.ManyUUID("Services", opts.ServiceIDs, 50)
	if err != nil {
		return nil, err
	}
	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("services", sqlutil.UUIDArray(opts.ServiceIDs)),
		sql.Named("until", opts.Until),
		sql.Named("since", opts.Since),
	}
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Record, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = new(SearchOptions)
	}

	data, err := (*renderData)(opts).Normalize()
	if err != nil {
		return nil, err
	}

	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	metrics := make([]Record, 0)
	var timeToAck sql.NullFloat64
	var timeToClose sql.NullFloat64
	for rows.Next() {
		var rec Record
		err := rows.Scan(&rec.ServiceID, &rec.ClosedAt, &rec.AlertCount, &timeToAck, &timeToClose)
		if err != nil {
			return nil, err
		}
		rec.TimeToClose = time.Duration(timeToClose.Float64 * float64(time.Second))
		rec.TimeToAck = time.Duration(timeToAck.Float64 * float64(time.Second))
		metrics = append(metrics, rec)
	}

	return metrics, nil
}
