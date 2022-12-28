package webhook

import (
	"context"
	"database/sql"
	"strconv"
	"text/template"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions allow filtering and paginating the list of timezones.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of timezone names to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name string `json:"n,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
        id,
		name,
        value
	FROM notification_channels nc
    WHERE true
	{{if .Omit}}
		AND not nc.name = any(:omit)
	{{end}}
	{{if .SearchStr}}
		AND (nc.name ILIKE :search)
	{{end}}
	{{if .After.Name}}
		AND lower(nc.name) > lower(:afterName)
	{{end}}
    AND nc.type = 'WEBHOOK'
	ORDER BY lower(nc.name)
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
		validate.Range("Omit", len(opts.Omit), 0, 50),
	)
	if opts.After.Name != "" {
		err = validate.Many(err, validate.Text("After.Name", opts.After.Name, 1, 255))
	}
	if err != nil {
		return nil, err
	}
	for i, name := range opts.Omit {
		err = validate.Many(err,
			validate.Text("Omit["+strconv.Itoa(i)+"]", name, 1, 255),
		)
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.StringArray(opts.Omit)),
		sql.Named("limit", opts.Limit),
	}
}

func (store *Store) Search(ctx context.Context, opts *SearchOptions) ([]Webhook, error) {
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
	rows, err := store.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Webhook
	var w Webhook
	for rows.Next() {
		err = rows.Scan(&w.ID, &w.Name, &w.Value)
		if err != nil {
			return nil, err
		}
		result = append(result, w)
	}

	return result, nil
}
