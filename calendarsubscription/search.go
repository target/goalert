package calendarsubscription

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
	"text/template"
)

// SearchOptions allow filtering and paginating the list of users.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of user IDs to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}
// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name string `json:"n,omitempty"`

}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		id, name, user_id, last_access, disabled
	FROM calendar_subscription cs 
	WHERE true
	{{if .Omit}}
		AND not id = any(:omit)
	{{end}}
	{{if .SearchStr}}
		AND cs.name ILIKE :search
	{{end}}
	{{if .After.Name}}
		AND lower(cs.name) > lower(:afterName)
	{{end}}
	ORDER BY lower(cs.name)
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
		validate.ManyUUID("Omit", opts.Omit, 50),
	)
	if opts.After.Name != "" {
		err = validate.Many(err, validate.Name("After.Name", opts.After.Name))
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
	}
}

func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]CalendarSubscription, error) {
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

	var result []CalendarSubscription
	var cs CalendarSubscription
	for rows.Next() {
		err = rows.Scan(&cs.ID, &cs.Name, &cs.UserID, &cs.LastAccess, &cs.Disabled)
		if err != nil {
			return nil, err
		}
		result = append(result, cs)
	}

	return result, nil
}



