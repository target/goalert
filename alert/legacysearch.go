package alert

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"
	"strings"

	"github.com/pkg/errors"
)

// LegacySearchOptions contains criteria for filtering and sorting alerts.
type LegacySearchOptions struct {
	// Search is matched case-insensitive against the alert summary, id and service name.
	Search string

	// ServiceID, if specified, will restrict alerts to those with a matching ServiceID.
	ServiceID string

	OmitTriggered bool
	OmitActive    bool
	OmitClosed    bool

	// Limit restricts the maximum number of rows returned. Default is 50.
	// Note: Limit is applied AFTER offset is taken into account.
	Limit int

	// Offset indicates the starting row of the returned results.
	Offset int

	// SortBy specifies the column to sort by. If anything other than ID,
	// ID is used as a secondary sort in descending (newest first) order.
	SortBy SortBy

	// SortDesc controls ascending or descending results of the primary sort (SortBy field).
	SortDesc bool

	//FavoriteServicesOnlyUserID, if populated, filters all those alerts which belong to this user's favorite services, if empty, it is ignored.
	FavoriteServicesOnlyUserID string
}

// SortBy describes the possible primary sort options for alerts.
type SortBy int

// Configurable sort columns.
const (
	SortByStatus SortBy = iota
	SortByID
	SortByCreatedTime
	SortBySummary
	SortByServiceName
)

// We need to escape any characters that have meaning for `ILIKE` in Postgres.
// https://www.postgresql.org/docs/8.3/static/functions-matching.html
var searchEscape = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// LegacySearch will return a list of matching alerts, up to Limit, and the total number of matches
// available.
func (db *DB) LegacySearch(ctx context.Context, opts *LegacySearchOptions) ([]Alert, int, error) {
	if opts == nil {
		opts = &LegacySearchOptions{}
	}

	userCheck := permission.User
	if opts.FavoriteServicesOnlyUserID != "" {
		userCheck = permission.MatchUser(opts.FavoriteServicesOnlyUserID)
	}
	err := permission.LimitCheckAny(ctx, permission.System, userCheck)
	if err != nil {
		return nil, 0, err
	}

	if opts.Limit == 0 {
		// default limit
		opts.Limit = 50
	}

	rawSearch := opts.Search
	if opts.Search != "" {
		// match any substring matching the literal (escaped) search string
		opts.Search = "%" + searchEscape.Replace(opts.Search) + "%"
	}
	err = validate.Many(
		validate.Range("Limit", opts.Limit, 15, 50),
		validate.Range("Offset", opts.Offset, 0, 1000000),
		validate.OneOf("SortBy", opts.SortBy, SortByID, SortByStatus, SortByCreatedTime, SortBySummary, SortByServiceName),
		validate.Text("Search", opts.Search, 0, 250),
	)
	if opts.FavoriteServicesOnlyUserID != "" {
		err = validate.Many(err, validate.UUID("FavoriteServicesOnlyUserID", opts.FavoriteServicesOnlyUserID))
	}
	if err != nil {
		return nil, 0, err
	}

	whereStr := `WHERE
		($1 = '' or cast(a.id as text) = $6 or svc.name ilike $1 or a.summary ilike $1) and
		($2 = '' or a.service_id = cast($2 as UUID)) and
		(
			($3 and a.status = 'triggered') or
			($4 and a.status = 'active') or
			($5 and a.status = 'closed')
		)
	`
	var buf strings.Builder
	buf.WriteString("ORDER BY\ncase when cast(a.id as text) = $6 then 0 else 1 end,\n")

	idSortType := "DESC"
	sortType := "ASC"
	if opts.SortDesc {
		sortType = "DESC"
	}
	switch opts.SortBy {
	case SortByStatus:
		buf.WriteString(fmt.Sprintf("a.status %s,\n", sortType))
	case SortByCreatedTime:
		buf.WriteString(fmt.Sprintf("a.created_at %s,\n", sortType))
	case SortBySummary:
		buf.WriteString(fmt.Sprintf("a.summary %s,\n", sortType))
	case SortByServiceName:
		buf.WriteString(fmt.Sprintf("svc.name %s,\n", sortType))
	case SortByID:
		if !opts.SortDesc {
			idSortType = "ASC"
		}
	}
	buf.WriteString(fmt.Sprintf("a.id %s\n", idSortType))

	orderStr := buf.String()

	queryArgs := []interface{}{
		opts.Search,
		opts.ServiceID,
		!opts.OmitTriggered,
		!opts.OmitActive,
		!opts.OmitClosed,
		rawSearch,
	}

	var favServiceOnlyFilter string
	// If FavoriteServicesOnlyFor userID is populated, use it for filtering
	if opts.FavoriteServicesOnlyUserID != "" {
		favServiceOnlyFilter = "JOIN user_favorites u ON u.tgt_service_id = a.service_id AND u.user_id = $7"
		queryArgs = append(queryArgs, opts.FavoriteServicesOnlyUserID)
	}

	totalQueryStr := `
		SELECT count(*)
		FROM alerts a
		JOIN services svc ON svc.id = a.service_id
	` + favServiceOnlyFilter + whereStr

	fetchQueryStr := fmt.Sprintf(`
		SELECT
			a.id,
			a.summary,
			a.details,
			a.service_id,
			a.source,
			a.status,
			a.created_at,
			a.dedup_key
		FROM alerts a
		JOIN services svc ON svc.id = a.service_id
		%s
		%s
		%s
		LIMIT %d
		OFFSET %d
	`, favServiceOnlyFilter, whereStr, orderStr, opts.Limit, opts.Offset)

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	var total int
	err = tx.QueryRowContext(ctx, totalQueryStr, queryArgs...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, "get total results")
	}
	rows, err := tx.QueryContext(ctx, fetchQueryStr, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var result []Alert
	for rows.Next() {
		var a Alert
		err = a.scanFrom(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, a)
	}

	return result, total, nil
}
