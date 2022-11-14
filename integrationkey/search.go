package integrationkey

import (
	"context"
	"database/sql"
	"text/template"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// InKeySearchOptions allow filtering and paginating the list of rotations.
type InKeySearchOptions struct {
	Search string `json:"s,omitempty"`
	After  string `json:"a,omitempty"`

	// Omit specifies a list of key ids to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

var intKeySearchTemplate = template.Must(template.New("integration-key-search").Parse(`
	SELECT DISTINCT
		key.id, key.name, key.type, key.service_id
	FROM integration_keys key
	WHERE true
	{{if .Omit}}
		AND not key.id = any(:omit)
	{{end}}
	{{if .Search}}
		AND (key.id::text ILIKE :search)
	{{end}}
	{{if .After}}
		lower(key.name) > lower(:after)
	{{end}}
	LIMIT {{.Limit}}
`))

type intKeyRenderData InKeySearchOptions

func (opts intKeyRenderData) Normalize() (*intKeyRenderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Search("Search", opts.Search),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.Range("Omit", len(opts.Omit), 0, 50),
	)

	if err != nil {
		return nil, err
	}

	return &opts, err
}

func (opts intKeyRenderData) SearchStr() string {
	if opts.Search == "" {
		return ""
	}

	return search.Escape(opts.Search) + "%"
}

func (opts intKeyRenderData) QueryArgs() []sql.NamedArg {

	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("after", opts.After),
		sql.Named("omit", sqlutil.StringArray(opts.Omit)),
	}
}

func (s *Store) Search(ctx context.Context, opts *InKeySearchOptions) ([]IntegrationKey, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &InKeySearchOptions{}
	}
	data, err := (*intKeyRenderData)(opts).Normalize()
	if err != nil {
		return nil, err
	}
	query, args, err := search.RenderQuery(ctx, intKeySearchTemplate, data)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []IntegrationKey
	for rows.Next() {
		var intKey IntegrationKey
		err = rows.Scan(&intKey.ID, &intKey.Name, &intKey.Type, &intKey.ServiceID)
		if err != nil {
			return nil, errors.Wrap(err, "scan row")
		}

		result = append(result, intKey)
	}

	return result, nil
}
