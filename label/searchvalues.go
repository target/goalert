package label

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

// ValueSearchOptions allow filtering and paginating the list of rotations.
type ValueSearchOptions struct {
	Key string `json:"k"`

	KeySearchOptions
}

var valueSearchTemplate = template.Must(template.New("value-search").Parse(`
	SELECT DISTINCT ON (lower(value), value) value
	FROM labels l
	WHERE key = :key
	{{if .Omit}}
		AND not value = any(:omit)
	{{end}}
	{{if .Search}}
		AND (l.value ILIKE :search)
	{{end}}
	{{if .After}}
		AND lower(l.value) > lower(:after) OR (lower(l.value) = lower(:after) AND l.value > :after)
	{{end}}
	ORDER BY lower(value), value
	LIMIT {{.Limit}}
`))

type valueRenderData ValueSearchOptions

func (opts valueRenderData) SearchValue() string {
	if opts.Search == "" {
		return ""
	}

	return "%" + search.Escape(opts.Search) + "%"
}
func (opts valueRenderData) Normalize() (*valueRenderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Search("Search", opts.Search),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.Range("Omit", len(opts.Omit), 0, 50),
		validate.Search("Key", opts.Key),
	)

	if opts.After != "" {
		err = validate.Many(err, validate.LabelValue("After", opts.After))
	}

	if err != nil {
		return nil, err
	}
	for i, value := range opts.Omit {
		err = validate.Many(err,
			validate.LabelKey("Omit["+strconv.Itoa(i)+"]", value),
		)
	}

	return &opts, err
}

func (opts valueRenderData) QueryArgs() []sql.NamedArg {

	return []sql.NamedArg{
		sql.Named("key", opts.Key),
		sql.Named("search", opts.SearchValue()),
		sql.Named("after", opts.After),
		sql.Named("omit", sqlutil.StringArray(opts.Omit)),
	}
}

func (db *DB) SearchValues(ctx context.Context, opts *ValueSearchOptions) ([]string, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &ValueSearchOptions{}
	}
	data, err := (*valueRenderData)(opts).Normalize()
	if err != nil {
		return nil, err
	}
	query, args, err := search.RenderQuery(ctx, valueSearchTemplate, data)
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

	var result []string
	for rows.Next() {
		var key string
		err = rows.Scan(&key)
		if err != nil {
			return nil, errors.Wrap(err, "scan row")
		}

		result = append(result, key)
	}

	return result, nil
}
