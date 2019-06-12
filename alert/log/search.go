package alertlog

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
	"text/template"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// SearchOptions contains criteria for filtering alert logs.
type SearchOptions struct {
	/// AlertID, if specified, will restrict alert logs to those with a matching AlertID.
	AlertID int

	// ServiceID, if specified, will restrict alert logs to those alerts which map to this particular ServiceID.
	ServiceID string

	// UserID, if specified, will restrict alert logs to those with events performed by the specified user.
	UserID string

	// IntegrationKeyID, if specified, will restrict alert logs to those with events authorized via the specified integration key.
	IntegrationKeyID string

	// Start will restrict alert logs to those which were created on or after this time.
	Start time.Time

	// End will restrict alert logs to those which were created before this time.
	End time.Time

	// Event, if specified, will restrict alert logs to those of the specified event type.
	Event Type

	// SortBy can be used to alter the primary sorting criteria. By default, results are ordered by timestamp as newest first.
	// Results will always have a secondary sort criteria of newest-events-first, unless SortByTimestamp is set and SortDesc is false.
	SortBy SortBy

	// SortDesc controls ascending or descending results of the primary sort (SortBy field).
	SortDesc bool

	// Offset indicates the starting row of the returned results.
	Offset int

	// Limit restricts the maximum number of rows returned. Default is 50.
	// Note: Limit is applied AFTER Offset is taken into account.
	Limit int `json:"-"`

	After SearchCursor `json:"a,omitempty"`
}

type SearchCursor struct {
	ID int `json:"i,omitempty"`
}

// SortBy describes the possible primary sort options for alert logs.
type SortBy int

// Configurable sort columns.
const (
	SortByTimestamp SortBy = iota
	SortByAlertID
	SortByEventType
	SortByUserName
)

var whereClause = `
		{{if .AlertID}}
		AND (a.alert_id = (:alertID::int))
		{{end}}	
		{{if .ServiceID}}
			AND (:serviceID = '' OR alerts.service_id = cast(:serviceID as UUID))
		{{end}}
		{{if .Start}}
			AND (coalesce(a.timestamp >= cast(:start as timestamp with time zone), true))
		{{end}}
		{{if .End}}
			AND (coalesce(a.timestamp < cast(:end as timestamp with time zone), true))
		{{end}}
		{{ if .Event }}
			AND (:event='' OR a.event = (:event::enum_alert_log_event))
		{{ end }}
		{{if .UserID}}
			AND (:userID = '' OR a.sub_user_id = cast(:userID as UUID))
		{{end}}
		{{if .IntegrationKeyID}}
			AND (:integrationKeyID = '' OR a.sub_integration_key_id = cast(:integrationKeyID as UUID))
		{{end}}
		{{if .After.ID}}
			AND (a.id > :afterID)
		{{end}}

		ORDER BY :orderBy 
		LIMIT {{.Limit}}
		OFFSET {{.Offset}}
`

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
			a.id, 
			a.alert_id,
			a.timestamp,
			a.event,
			a.message,
			a.sub_type,
			a.sub_user_id,
			u.name,
			a.sub_integration_key_id,
			i.name,
			a.sub_hb_monitor_id,
			hb.name,
			a.sub_channel_id,
			nc.name,
			a.sub_classifier,
			a.meta
		FROM alert_logs a
		LEFT JOIN alerts ON alerts.id = a.alert_id
		LEFT JOIN users u ON u.id = a.sub_user_id
		LEFT JOIN integration_keys i ON i.id = a.sub_integration_key_id
		LEFT JOIN heartbeat_monitors hb ON hb.id = a.sub_hb_monitor_id 
		LEFT JOIN notification_channels nc ON nc.id = a.sub_channel_id
		WHERE true	
  ` + whereClause))

var totalTemplate = template.Must(template.New("totalQuery").Parse(`
	SELECT count(*)
		FROM alert_logs a
		JOIN alerts ON alerts.id = a.alert_id
		WHERE true
  ` + whereClause))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	var buf bytes.Buffer

	idSortType := "DESC"
	// sortType only applies to user-specified parameter
	sortType := "ASC"
	if opts.SortDesc {
		sortType = "DESC"
	}

	switch opts.SortBy {
	case SortByTimestamp:
		if !opts.SortDesc { // if SortDesc is false
			idSortType = "ASC"
		}
	case SortByAlertID:
		buf.WriteString(fmt.Sprintf("a.alert_id %s,\n", sortType))
	case SortByEventType:
		buf.WriteString(fmt.Sprintf("cast(a.event as text) %s,\n", sortType))
	case SortByUserName:
		buf.WriteString(fmt.Sprintf("u.name %s,\n", sortType))
	}

	// idSortType is applied to both timestamp and id
	buf.WriteString(fmt.Sprintf("a.timestamp %s,\n", idSortType))
	buf.WriteString(fmt.Sprintf("a.id %s\n", idSortType))

	return buf.String()
}

func (opts renderData) StartTime() pq.NullTime {
	var start pq.NullTime
	if !opts.Start.IsZero() {
		start.Valid = true
		start.Time = opts.Start
	}
	return start
}

func (opts renderData) EndTime() pq.NullTime {
	var end pq.NullTime
	if !opts.End.IsZero() {
		end.Valid = true
		end.Time = opts.End
	}
	return end
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("start", opts.StartTime()),
		sql.Named("end", opts.EndTime()),
		sql.Named("serviceID", opts.ServiceID),
		sql.Named("alertID", opts.AlertID),
		sql.Named("event", opts.Event),
		sql.Named("userID", opts.UserID),
		sql.Named("integrationKeyID", opts.IntegrationKeyID),
		sql.Named("orderBy", opts.OrderBy()),
		sql.Named("afterID", opts.After.ID),
	}
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	var err error
	if opts.ServiceID != "" {
		err = validate.Many(err, validate.UUID("ServiceID", opts.ServiceID))
	}

	if opts.UserID != "" {
		err = validate.Many(err, validate.UUID("UserID", opts.UserID))
	}

	if opts.IntegrationKeyID != "" {
		err = validate.Many(err, validate.UUID("IntegrationKeyID", opts.IntegrationKeyID))
	}

	err = validate.Many(err, validate.OneOf("SortBy", opts.SortBy,
		SortByAlertID,
		SortByEventType,
		SortByTimestamp,
		SortByUserName,
	))

	err = validate.Many(err, validate.Range("Limit", opts.Limit, 1, 50),
		validate.Range("Offset", opts.Offset, 0, 1000000))

	if err != nil {
		return nil, err
	}

	return &opts, nil
}

// Search will return a list of matching log entries and the total number of matches available.
func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]Entry, int, error) {
	var total int
	if opts == nil {
		opts = &SearchOptions{}
	}

	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User, permission.System)
	if err != nil {
		return nil, 0, err
	}

	if opts.Limit == 0 {
		// default limit
		opts.Limit = 25
	}

	data, err := (*renderData)(opts).Normalize()
	if err != nil {
		return nil, -1, err
	}

	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
	if err != nil {
		return nil, -1, errors.Wrap(err, "render query")
	}

	rows, err := db.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, -1, err
	}
	defer rows.Close()

	var result []rawEntry

	for rows.Next() {
		var r rawEntry
		err = r.scanWith(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, r)
	}
	var logs []Entry
	for _, e := range result {
		logs = append(logs, e)
	}

	// Getting total number of results
	query, args, err = search.RenderQuery(ctx, totalTemplate, data)
	if err != nil {
		return nil, -1, errors.Wrap(err, "render query")
	}

	err = db.db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return nil, -1, err
	}
	defer rows.Close()

	return logs, total, nil
}