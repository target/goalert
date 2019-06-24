package alertlog

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// LegacySearchOptions contains criteria for filtering alert logs. At a minimum, at least one of AlertID or ServiceID must be specified.
type LegacySearchOptions struct {
	// AlertID, if specified, will restrict alert logs to those with a matching AlertID.
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

	// Limit restricts the maximum number of rows returned. Default is 25. Maximum is 50.
	// Note: Limit is applied AFTER Offset is taken into account.
	Limit int
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

// LegacySearch will return a list of matching log entries.
func (db *DB) LegacySearch(ctx context.Context, opts *LegacySearchOptions) ([]Entry, int, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User, permission.System)
	if err != nil {
		return nil, 0, err
	}

	if opts.Limit == 0 {
		// default limit
		opts.Limit = 25
	}

	if opts.ServiceID == "" && opts.AlertID == 0 {
		err = validation.NewFieldError("LegacySearchOptions", "One of AlertID or ServiceID must be specified")
	}

	err = validate.Many(
		err,
		validate.Range("Limit", opts.Limit, 1, search.MaxResults),
		validate.Range("Offset", opts.Offset, 0, 1000000),
		validate.OneOf("SortBy", opts.SortBy,
			SortByAlertID,
			SortByEventType,
			SortByTimestamp,
			SortByUserName),
	)
	if err != nil {
		return nil, 0, err
	}

	var buf bytes.Buffer
	idSortType := "DESC"
	// sortType only applies to user-specified parameter
	sortType := "ASC"
	if opts.SortDesc {
		sortType = "DESC"
	}

	buf.WriteString("ORDER BY ")

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

	orderStr := buf.String()
	// Refer to https://github.com/jackc/pgx/issues/281 on why to include a typecast before comparing to null
	whereStr := `WHERE 
	($1 = '0' or a.alert_id = $1 ::int) and 
	($2 = '' or alerts.service_id = cast($2 as UUID)) and
	(coalesce(a.timestamp >= cast($3 as timestamp with time zone), true)) and 
	(coalesce(a.timestamp < cast($4 as timestamp with time zone), true)) and
	($5 = '' or a.event = $5::enum_alert_log_event)and
	($6 = '' or a.sub_user_id = cast($6 as UUID)) and 
	($7 = '' or a.sub_integration_key_id = cast($7 as UUID))`

	fetchQueryStr := fmt.Sprintf(`
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
		%s
		%s
		LIMIT %d
		OFFSET %d
	`, whereStr, orderStr, opts.Limit, opts.Offset)

	totalQueryStr := `
		SELECT count(*)
		FROM alert_logs a
		JOIN alerts ON alerts.id = a.alert_id
	` + whereStr

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var start, end pq.NullTime
	if !opts.Start.IsZero() {
		start.Valid = true
		start.Time = opts.Start
	}
	if !opts.End.IsZero() {
		end.Valid = true
		end.Time = opts.End
	}

	var total int
	err = tx.QueryRowContext(ctx, totalQueryStr,
		opts.AlertID,
		opts.ServiceID,
		start,
		end,
		opts.Event,
		opts.UserID,
		opts.IntegrationKeyID,
	).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, "get total results")
	}

	rows, err := tx.QueryContext(ctx, fetchQueryStr,
		opts.AlertID,
		opts.ServiceID,
		start,
		end,
		opts.Event,
		opts.UserID,
		opts.IntegrationKeyID,
	)
	if err != nil {
		return nil, 0, err
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

	return logs, total, nil

}
