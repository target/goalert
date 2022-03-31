package alertmetrics

import (
	"context"
	"database/sql"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/alert"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions contains criteria for filtering and sorting alert metrics.
type SearchOptions struct {
	// ServiceIDs, if specified, will restrict alert metrics to those with a matching ServiceID on IDs, if valid.
	ServiceIDs []string `json:"v,omitempty"`

	// LowerBound will omit any alert metrics created any time before the provided time.
	LowerBound time.Time `json:"n,omitempty"`

	// UpperBound will only include alert metrics that were created before the provided time.
	UpperBound time.Time `json:"b,omitempty"`
}

var searchTemplate = template.Must(template.New("alert-metrics-search").Funcs(search.Helpers()).Parse(`
	SELECT
		am.alert_id,
		am.service_id,
		a.created_at
	FROM alert_metrics am
	JOIN alerts a
	ON a.id = am.alert_id
	WHERE true
	{{if .ServiceIDs}}
		AND am.service_id = any(:services)
	{{end}}
	{{ if not .UpperBound.IsZero }}
		AND a.created_at < :upperBound
	{{ end }}
	{{ if not .LowerBound.IsZero }}
		AND a.created_at >= :lowerBound
	{{ end }}
	AND a.status = 'closed'
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
		sql.Named("upperBound", opts.UpperBound),
		sql.Named("lowerBound", opts.LowerBound),
	}
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]alert.DataPoint, error) {
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

	metrics := make([]alert.DataPoint, 0)
	for rows.Next() {
		var dp alert.DataPoint
		err := rows.Scan(&dp.AlertID, &dp.ServiceID, &dp.Timestamp)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, dp)
	}

	return metrics, nil
}
