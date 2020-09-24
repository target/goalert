package alert

import (
	"context"
	"database/sql"
	"strconv"
	"text/template"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SortMode indicates the mode of sorting for alerts.
type SortMode int

const (
	// SortModeStatusID will sort by status priority (unacked, then acked, then closed) followed by ID (newest/highest first)
	SortModeStatusID SortMode = iota

	// SortModeDateID will sort alerts by date newest first, falling back to ID (newest/highest first)
	SortModeDateID

	// SortModeDateIDReverse will sort alerts by date oldest first, falling back to ID (oldest/lowest first)
	SortModeDateIDReverse
)

// SearchOptions contains criteria for filtering and sorting alerts.
type SearchOptions struct {
	// Search is matched case-insensitive against the alert summary, id and service name.
	Search string `json:"s,omitempty"`

	// Status, if specified, will restrict alerts to those with a matching status.
	Status []Status `json:"t,omitempty"`

	// ServiceFilter, if specified, will restrict alerts to those with a matching ServiceID on IDs, if valid.
	ServiceFilter IDFilter `json:"v,omitempty"`

	After SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of alert IDs to exclude from the results.
	Omit []int `json:"o,omitempty"`

	// NotifiedUserID will include all alerts the specified user has been
	// notified for to the results.
	NotifiedUserID string `json:"e,omitempty"`

	// Limit restricts the maximum number of rows returned. Default is 50.
	// Note: Limit is applied AFTER AfterID is taken into account.
	Limit int `json:"-"`

	// Sort allows customizing the sort method.
	Sort SortMode `json:"z,omitempty"`

	// NotBefore will omit any alerts created any time before the provided time.
	NotBefore time.Time `json:"n,omitempty"`

	// Before will only include alerts that were created before the provided time.
	Before time.Time `json:"b,omitempty"`
}

type IDFilter struct {
	Valid bool     `json:"v,omitempty"`
	IDs   []string `json:"i,omitempty"`
}

type SearchCursor struct {
	ID      int       `json:"i,omitempty"`
	Status  Status    `json:"s,omitempty"`
	Created time.Time `json:"c,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		a.id,
		a.summary,
		a.details,
		a.service_id,
		a.source,
		a.status,
		created_at,
		a.dedup_key
	FROM alerts a
	{{ if .Search }}
		JOIN services svc ON svc.id = a.service_id
	{{ end }}
	WHERE true
	{{ if .Omit }}
		AND not a.id = any(:omit)
	{{ end }}
	{{ if .Search }}
		AND (
			a.id = :searchID OR
			a.summary ilike :search OR
			svc.name ilike :search
		)
	{{ end }}
	{{ if .Status }}
		AND a.status = any(:status::enum_alert_status[])
	{{ end }}
	{{ if .ServiceFilter.Valid }}
		AND (a.service_id = any(:services)
			{{ if .NotifiedUserID }}
				OR a.id = any(select alert_id from alert_logs where event in ('notification_sent', 'no_notification_sent') and sub_user_id = :notifiedUserID)
			{{ end }}
		)
	{{ end }}
	{{ if not .Before.IsZero }}
		AND a.created_at < :beforeTime
	{{ end }}
	{{ if not .NotBefore.IsZero }}
		AND a.created_at >= :notBeforeTime
	{{ end }}
	{{ if .After.ID }}
		AND (
			{{ if eq .Sort 1 }}
				a.created_at < :afterCreated OR
				(a.created = :afterCreated AND a.id < :afterID)
			{{ else if eq .Sort 2}}
				a.created_at > :afterCreated OR
				(a.created_at = :afterCreated AND a.id > :afterID)
			{{ else }}
				a.status > :afterStatus::enum_alert_status OR
				(a.status = :afterStatus::enum_alert_status AND a.id < :afterID)
			{{ end }}
		)
	{{ end }}
	ORDER BY {{.SortStr}}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) SortStr() string {
	switch opts.Sort {
	case SortModeDateID:
		return "created_at DESC, id DESC"
	case SortModeDateIDReverse:
		return "created_at, id"
	}

	// SortModeStatusID
	return "status, id DESC"
}

func (opts renderData) SearchStr() string {
	if opts.Search == "" {
		return ""
	}

	return "%" + search.Escape(opts.Search) + "%"
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Search("Search", opts.Search),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.Range("Status", len(opts.Status), 0, 3),
		validate.ManyUUID("Services", opts.ServiceFilter.IDs, 50),
		validate.Range("Omit", len(opts.Omit), 0, 50),
		validate.OneOf("Sort", opts.Sort, SortModeStatusID, SortModeDateID, SortModeDateIDReverse),
	)
	if opts.After.Status != "" {
		err = validate.Many(err, validate.OneOf("After.Status", opts.After.Status, StatusTriggered, StatusActive, StatusClosed))
	}
	if err != nil {
		return nil, err
	}

	for i, stat := range opts.Status {
		err = validate.OneOf("Status["+strconv.Itoa(i)+"]", stat, StatusTriggered, StatusActive, StatusClosed)
		if err != nil {
			return nil, err
		}
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	var searchID sql.NullInt64
	if i, err := strconv.ParseInt(opts.Search, 10, 64); err == nil {
		searchID.Valid = true
		searchID.Int64 = i
	}

	stat := make(sqlutil.StringArray, len(opts.Status))
	for i := range opts.Status {
		stat[i] = string(opts.Status[i])
	}

	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("searchID", searchID),
		sql.Named("status", stat),
		sql.Named("services", sqlutil.UUIDArray(opts.ServiceFilter.IDs)),
		sql.Named("afterID", opts.After.ID),
		sql.Named("afterStatus", opts.After.Status),
		sql.Named("afterCreated", opts.After.Created),
		sql.Named("omit", sqlutil.IntArray(opts.Omit)),
		sql.Named("notifiedUserID", opts.NotifiedUserID),
		sql.Named("beforeTime", opts.Before),
		sql.Named("notBeforeTime", opts.NotBefore),
	}
}

func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]Alert, error) {
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

	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}
	defer rows.Close()

	alerts := make([]Alert, 0, opts.Limit)

	for rows.Next() {
		var a Alert
		err = errors.Wrap(a.scanFrom(rows.Scan), "scan")
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}
