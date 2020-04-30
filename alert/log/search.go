package alertlog

import (
	"context"
	"database/sql"
	"text/template"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SearchOptions contains criteria for filtering alert logs.
type SearchOptions struct {
	// FilterAlertIDs restricts the log entries belonging to specific alertIDs only.
	FilterAlertIDs []int `json:"f"`

	// Limit restricts the maximum number of rows returned. Default is 15.
	Limit int `json:"-"`

	After SearchCursor `json:"a,omitempty"`
}

type SearchCursor struct {
	ID int `json:"i,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		log.id, 
		log.alert_id,
		log.timestamp,
		log.event,
		log.message,
		log.sub_type,
		log.sub_user_id,
		u.name,
		log.sub_integration_key_id,
		i.name,
		log.sub_hb_monitor_id,
		hb.name,
		log.sub_channel_id,
		nc.name,
		log.sub_classifier,
		log.meta,
		om.last_status,
		om.status_details
	FROM alert_logs log
	LEFT JOIN users u ON u.id = log.sub_user_id
	LEFT JOIN integration_keys i ON i.id = log.sub_integration_key_id
	LEFT JOIN heartbeat_monitors hb ON hb.id = log.sub_hb_monitor_id 
	LEFT JOIN notification_channels nc ON nc.id = log.sub_channel_id
	LEFT JOIN outgoing_messages om ON om.id = (log.meta->>'MessageID')::uuid
	WHERE TRUE
	{{- if .FilterAlertIDs}}
		AND log.alert_id = ANY(:alertIDs)
	{{- end}}
	{{- if .After.ID}}
		AND (log.id < :afterID)
	{{- end}}
	ORDER BY log.id DESC
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Range("FilterAlertIDs", len(opts.FilterAlertIDs), 0, 50),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
	)
	if err != nil {
		return nil, err
	}

	return &opts, nil
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("afterID", opts.After.ID),
		sql.Named("alertIDs", sqlutil.IntArray(opts.FilterAlertIDs)),
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

	var result []Entry
	for rows.Next() {
		var r Entry
		err = r.scanWith(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil

}
