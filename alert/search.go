package alert

import (
	"context"
	"database/sql"
	"strconv"
	"text/template"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SearchOptions contains criteria for filtering and sorting alerts.
type SearchOptions struct {
	// Search is matched case-insensitive against the alert summary, id and service name.
	Search string `json:"s,omitempty"`

	// Status, if specified, will restrict alerts to those with a matching status.
	Status []Status `json:"t,omitempty"`

	// Services, if specified, will restrict alerts to those with a matching ServiceID.
	Services []string `json:"v,omitempty"`

	After SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of alert IDs to exclude from the results.
	Omit []int `json:"o,omitempty"`

	// NotifiedUserID will include all alerts the specified user has been
	// notified for to the results.
	NotifiedUserID string `json:"e,omitempty"`

	// Limit restricts the maximum number of rows returned. Default is 50.
	// Note: Limit is applied AFTER AfterID is taken into account.
	Limit int `json:"-"`
}

type SearchCursor struct {
	ID     int    `json:"i,omitempty"`
	Status Status `json:"s,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	{{ if .NotifiedUserID }}
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
		JOIN alert_logs al ON al.alert_id = a.id AND al.event = 'notification_sent'
		WHERE al.sub_user_id = :currentUserID
		UNION
	{{ end }}

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
	{{ if .Services }}
		AND a.service_id = any(:services)
	{{ end }}
	{{ if .After.ID }}
		AND (
			a.status > :afterStatus::enum_alert_status OR
			(a.status = :afterStatus::enum_alert_status AND a.id < :afterID)
		)
	{{ end }}
	ORDER BY status, id DESC
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

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
		validate.ManyUUID("Services", opts.Services, 50),
		validate.Range("Omit", len(opts.Omit), 0, 50),
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
		sql.Named("services", sqlutil.UUIDArray(opts.Services)),
		sql.Named("afterID", opts.After.ID),
		sql.Named("afterStatus", opts.After.Status),
		sql.Named("omit", sqlutil.IntArray(opts.Omit)),
		sql.Named("currentUserID", opts.NotifiedUserID),
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
