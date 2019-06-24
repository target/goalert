package alertlog

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"text/template"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// SearchOptions contains criteria for filtering alert logs.
type SearchOptions struct {
	// FilterAlertIDs restricts the log entries belonging to specific alertIDs only.
	FilterAlertIDs []int

	// Limit restricts the maximum number of rows returned. Default is 15.
	Limit int

	After SearchCursor `json:"a,omitempty"`
}

type SearchCursor struct {
	ID int `json:"i,omitempty"`
}

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
	{{if .FilterAlertIDs }}
		LEFT JOIN alerts ON
			alerts.id = a.alert_id
	{{end}}
	LEFT JOIN users u ON u.id = a.sub_user_id
	LEFT JOIN integration_keys i ON i.id = a.sub_integration_key_id
	LEFT JOIN heartbeat_monitors hb ON hb.id = a.sub_hb_monitor_id 
	LEFT JOIN notification_channels nc ON nc.id = a.sub_channel_id
	WHERE TRUE
	{{if .FilterAlertIDs}}
		AND a.alert_id = ANY(:alertIDs)
	{{end}}
	{{- if .After.ID}}
		AND (a.id < :afterID)
	{{- end}}
	ORDER BY {{ .OrderBy }}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	return "a.id DESC"
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	var err error
	if len(opts.FilterAlertIDs) <= 0 {
		err = validation.NewFieldError("SearchOptions", "FilterAlertIDs must be specified")
	}

	err = validate.Many(
		err,
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
	)
	if err != nil {
		return nil, err
	}

	return &opts, nil
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	alertIDs := make(pq.Int64Array, len(opts.FilterAlertIDs))
	for i := range opts.FilterAlertIDs {
		alertIDs[i] = int64(opts.FilterAlertIDs[i])
	}

	return []sql.NamedArg{
		sql.Named("afterID", opts.After.ID),
		sql.Named("alertIDs", alertIDs),
	}
}

// Search will return a list of matching log entries
func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]Entry, error) {
	if opts == nil {
		opts = &SearchOptions{}
	}

	err := permission.LimitCheckAny(ctx, permission.System, permission.Admin, permission.User)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	defer rows.Close()

	var result []rawEntry
	for rows.Next() {
		var r rawEntry
		err = r.scanWith(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	var logs []Entry
	for _, e := range result {
		logs = append(logs, e)
	}

	return logs, nil

}
