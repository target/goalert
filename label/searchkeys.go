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

// KeySearchOptions allow filtering and paginating the list of rotations.
type KeySearchOptions struct {
	Search string `json:"s,omitempty"`
	After  string `json:"a,omitempty"`

	// Omit specifies a list of key names to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

var keySearchTemplate = template.Must(template.New("key-search").Parse(`
	SELECT DISTINCT ON (lower(key), key) key
	FROM labels l
	WHERE true
	{{if .Omit}}
		AND not key = any(:omit)
	{{end}}
	{{if .Search}}
		AND (l.key ILIKE :search)
	{{end}}
	{{if .After}}
		AND lower(l.key) > lower(:after) OR (lower(l.key) = lower(:after) AND l.key > :after)
	{{end}}
	ORDER BY lower(key), key
	LIMIT {{.Limit}}
`))

type keyRenderData KeySearchOptions

func (opts keyRenderData) SearchValue() string {
	if opts.Search == "" {
		return ""
	}

	return "%" + search.Escape(opts.Search) + "%"
}

func (opts keyRenderData) Normalize() (*keyRenderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Search("Search", opts.Search),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.Range("Omit", len(opts.Omit), 0, 50),
	)

	if opts.After != "" {
		err = validate.Many(err, validate.LabelKey("After", opts.After))
	}

	if err != nil {
		return nil, err
	}
	for i, key := range opts.Omit {
		err = validate.Many(err,
			validate.LabelKey("Omit["+strconv.Itoa(i)+"]", key),
		)
	}

	return &opts, err
}

func (opts keyRenderData) QueryArgs() []sql.NamedArg {

	return []sql.NamedArg{
		sql.Named("search", opts.SearchValue()),
		sql.Named("after", opts.After),
		sql.Named("omit", sqlutil.StringArray(opts.Omit)),
	}
}

func (db *DB) SearchKeys(ctx context.Context, opts *KeySearchOptions) ([]string, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &KeySearchOptions{}
	}
	data, err := (*keyRenderData)(opts).Normalize()
	if err != nil {
		return nil, err
	}
	query, args, err := search.RenderQuery(ctx, keySearchTemplate, data)
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
