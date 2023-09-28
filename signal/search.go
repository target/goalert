package signal

import (
	"context"
	"database/sql"
	"text/template"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SortMode indicates the mode of sorting for signals.
type SortMode int

const (
	// SortModeDateID will sort signals by date newest first, falling back to ID (newest/highest first)
	SortModeDateID SortMode = iota

	// SortModeDateIDReverse will sort signals by date oldest first, falling back to ID (oldest/lowest first)
	SortModeDateIDReverse
)

// SearchOptions contains criteria for filtering and sorting signals.
type SearchOptions struct {

	// ServiceFilter, if specified, will restrict signals to those with a matching ServiceID on IDs, if valid.
	ServiceFilter IDFilter `json:"v,omitempty"`

	// ServiceRuleFilter, if specified, will restrict signals to those with a matching ServiceRuleID on IDs, if valid.
	ServiceRuleFilter IDFilter `json:"r,omitempty"`

	After SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of signal IDs to exclude from the results.
	Omit []int `json:"o,omitempty"`

	// Limit restricts the maximum number of rows returned. Default is 50.
	// Note: Limit is applied AFTER AfterID is taken into account.
	Limit int `json:"-"`

	// Sort allows customizing the sort method.
	Sort SortMode `json:"z,omitempty"`

	// NotBefore will omit any signals with a timestamp before the provided time.
	NotBefore time.Time `json:"n,omitempty"`

	// Before will only include signals with a timestamp before the provided time.
	Before time.Time `json:"b,omitempty"`
}

type IDFilter struct {
	Valid bool     `json:"v,omitempty"`
	IDs   []string `json:"i,omitempty"`
}

type SearchCursor struct {
	ID        int       `json:"i,omitempty"`
	Timestamp time.Time `json:"c,omitempty"`
}

var searchTemplate = template.Must(template.New("signal-search").Funcs(search.Helpers()).Parse(`
	SELECT
		s.id,
		s.service_rule_id,
		s.service_id,
		s.outgoing_payload,
		s.scheduled,
		s.timestamp
	FROM signals s
	WHERE true
	{{ if .Omit }}
		AND not s.id = any(:omit)
	{{ end }}
	{{ if .ServiceFilter.Valid }}
		AND (s.service_id = any(:services))
	{{ end }}
	{{ if .ServiceRuleFilter.Valid }}
		AND (s.service_rule_id = any(:serviceRules))
	{{ end }}
	{{ if not .Before.IsZero }}
		AND s.timestamp < :beforeTime
	{{ end }}
	{{ if not .NotBefore.IsZero }}
		AND s.timestamp >= :notBeforeTime
	{{ end }}
	{{ if .After.ID }}
		AND (
			{{ if eq .Sort 1 }}
				s.timestamp < :afterTimestamp OR
				(s.timestamp = :afterTimestamp AND s.id < :afterID)
			{{ else if eq .Sort 2}}
				s.timestamp > :afterTimestamp OR
				(s.timestamp = :afterTimestamp AND s.id > :afterID)
			{{ end }}
		)
	{{ end }}
	ORDER BY {{.SortStr}}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) SortStr() string {
	if opts.Sort == SortModeDateIDReverse {
		return "timestamp, id"
	}

	// SortModeDateID
	return "timestamp DESC, id DESC"
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Range("Limit", opts.Limit, 0, 1001),
		validate.ManyUUID("Services", opts.ServiceFilter.IDs, 50),
		validate.ManyUUID("ServiceRules", opts.ServiceRuleFilter.IDs, 50),
		validate.Range("Omit", len(opts.Omit), 0, 50),
		validate.OneOf("Sort", opts.Sort, SortModeDateID, SortModeDateIDReverse),
	)
	if err != nil {
		return nil, err
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	var searchID sql.NullInt64

	return []sql.NamedArg{
		sql.Named("searchID", searchID),
		sql.Named("services", sqlutil.UUIDArray(opts.ServiceFilter.IDs)),
		sql.Named("serviceRules", sqlutil.UUIDArray(opts.ServiceRuleFilter.IDs)),
		sql.Named("afterID", opts.After.ID),
		sql.Named("afterTimestamp", opts.After.Timestamp),
		sql.Named("omit", sqlutil.IntArray(opts.Omit)),
		sql.Named("beforeTime", opts.Before),
		sql.Named("notBeforeTime", opts.NotBefore),
	}
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Signal, error) {
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

	signals := make([]Signal, 0, opts.Limit)

	for rows.Next() {
		var a Signal
		err = errors.Wrap(a.scanFrom(rows.Scan), "scan")
		if err != nil {
			return nil, err
		}
		signals = append(signals, a)
	}

	return signals, nil
}
