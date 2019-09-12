package override

import (
	"context"
	"database/sql"
	"github.com/target/goalert/util/sqlutil"
	"text/template"
	"time"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SearchOptions allow filtering and paginating the list of rotations.
type SearchOptions struct {
	After SearchCursor `json:"a,omitempty"`
	Limit int          `json:"-"`

	// Omit specifies a list of override IDs to exclude from the results.
	Omit []string

	ScheduleID string `json:"d,omitempty"`

	AddUserIDs    []string `json:"u,omitempty"`
	RemoveUserIDs []string `json:"r,omitempty"`
	AnyUserIDs    []string `json:"n,omitempty"`

	Start time.Time `json:"t,omitempty"`
	End   time.Time `json:"e,omitempty"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	ID string `json:"i,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	{{if .After.ID}}
	WITH after AS (
		SELECT id, start_time, end_time
		FROM user_overrides
		WHERE id = :afterID
	)
	{{end}}
	SELECT
		o.id, o.start_time, o.end_time, add_user_id, remove_user_id, tgt_schedule_id
	FROM user_overrides o
	{{if .After.ID}}
	JOIN after ON true
	{{end}}
	WHERE true
	{{if .Omit}}
		AND not o.id = any(:omit)
	{{end}}
	{{if .ScheduleID}}
		AND tgt_schedule_id = :scheduleID
	{{end}}
	{{if .AnyUserIDs}}
		AND (add_user_id = any(:anyUserIDs) OR remove_user_id = any(:anyUserIDs))
	{{end}}
	{{if .AddUserIDs}}
		AND add_user_id = any(:addUserIDs)
	{{end}}
	{{if .RemoveUserIDs}}
		AND remove_user_id = any(:removeUserIDs)
	{{end}}
	{{if not .Start.IsZero}}
		AND o.end_time > :startTime
	{{end}}
	{{if not .End.IsZero}}
		AND o.start_time <= :endTime
	{{end}}
	{{if .After.ID}}
		AND (
			o.start_time > after.start_time OR (
				o.start_time = after.start_time AND
				o.end_time > after.end_time
			) OR (
				o.start_time = after.start_time AND
				o.end_time = after.end_time AND
				o.id > after.id
			)
		)
	{{end}}
	ORDER BY o.start_time, o.end_time, o.id
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.ManyUUID("AddUserIDs", opts.AddUserIDs, 10),
		validate.ManyUUID("RemoveUserIDs", opts.RemoveUserIDs, 10),
		validate.ManyUUID("AnyUserIDs", opts.RemoveUserIDs, 10),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.ManyUUID("Omit", opts.Omit, 50),
	)
	if opts.ScheduleID != "" {
		err = validate.Many(err, validate.UUID("ScheduleID", opts.ScheduleID))
	}
	if opts.After.ID != "" {
		err = validate.Many(err, validate.UUID("After.ID", opts.After.ID))
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("afterID", opts.After.ID),
		sql.Named("scheduleID", opts.ScheduleID),
		sql.Named("startTime", opts.Start),
		sql.Named("endTime", opts.End),
		sql.Named("addUserIDs", sqlutil.UUIDArray(opts.AddUserIDs)),
		sql.Named("removeUserIDs", sqlutil.UUIDArray(opts.RemoveUserIDs)),
		sql.Named("anyUserIDs", sqlutil.UUIDArray(opts.AnyUserIDs)),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
	}
}

func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]UserOverride, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &SearchOptions{}
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
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []UserOverride
	var u UserOverride
	var add, rem, schedID sql.NullString
	for rows.Next() {
		err = rows.Scan(&u.ID, &u.Start, &u.End, &add, &rem, &schedID)
		if err != nil {
			return nil, err
		}
		u.AddUserID = add.String
		u.RemoveUserID = rem.String
		if schedID.Valid {
			u.Target = assignment.ScheduleTarget(schedID.String)
		}
		result = append(result, u)
	}

	return result, nil
}
