package schedule

import (
	"context"
	"database/sql"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
	"text/template"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// SearchOptions allow filtering and paginating the list of schedules.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of schedule IDs to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name string `json:"n,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		id, name, description, time_zone
	FROM schedules sched
	WHERE true
	{{if .Omit}}
		AND not id = any(:omit)
	{{end}}
	{{if .SearchStr}}
		AND (sched.name ILIKE :search OR sched.description ILIKE :search)
	{{end}}
	{{if .After.Name}}
		AND lower(sched.name) > lower(:afterName)
	{{end}}
	ORDER BY lower(sched.name)
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
		validate.Text("Search", opts.Search, 0, search.MaxQueryLen),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.ManyUUID("Omit", opts.Omit, 50),
	)
	if opts.After.Name != "" {
		err = validate.Many(err, validate.IDName("After.Name", opts.After.Name))
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", pq.StringArray(opts.Omit)),
	}
}

func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]Schedule, error) {
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

	var result []Schedule
	var s Schedule
	var tz string
	for rows.Next() {
		err = rows.Scan(&s.ID, &s.Name, &s.Description, &tz)
		if err != nil {
			return nil, err
		}
		loc, err := util.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
		s.TimeZone = loc
		result = append(result, s)
	}

	return result, nil
}
