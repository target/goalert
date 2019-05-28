package label

import (
	"context"
	"database/sql"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
	"strconv"
	"strings"
	"text/template"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

// SearchOptions allow filtering and paginating the list of rotations.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of key names to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`

	UniqueKeys bool `json:"u,omitempty"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Key        string                `json:"k,omitempty"`
	TargetID   string                `json:"t,omitempty"`
	TargetType assignment.TargetType `json:"y,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT{{if .UniqueKeys}} distinct on (lower(key)){{end}}
		key, value, tgt_service_id
	FROM labels l
	WHERE true
	{{if .Omit}}
		AND not key = any(:omit)
	{{end}}
	{{if .KeySearch}}
		AND (l.key ILIKE :keySearch)
	{{end}}
	{{if .ValueSearch}}
		AND ({{if .ValueNegate}}NOT {{end}}l.value ILIKE :valueSearch)
	{{end}}
	{{if .After.Key}}
		AND (lower(l.key) > lower(:afterKey) AND l.tgt_service_id > :afterServiceID)
	{{end}}
	ORDER BY lower(key), tgt_service_id
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) ValueNegate() bool {
	idx := strings.IndexRune(opts.Search, '=')
	return idx > 0 && opts.Search[idx-1] == '!'
}
func (opts renderData) KeySearch() string {
	if opts.Search == "" {
		return ""
	}

	idx := strings.IndexRune(opts.Search, '=')
	if idx != -1 {
		s := search.Escape(strings.TrimSuffix(opts.Search[:idx], "!"))
		if s == "*" {
			return ""
		}
		// Equal sign denotes exact match, however
		// up to 2 wildcards are  supported via '*'.
		return strings.Replace(s, "*", "%", 2)
	}

	return "%" + search.Escape(opts.Search) + "%"
}
func (opts renderData) ValueSearch() string {
	if opts.Search == "" {
		return ""
	}

	idx := strings.IndexRune(opts.Search, '=')
	if idx == -1 {
		return ""
	}
	s := search.Escape(opts.Search[idx+1:])
	if s == "*" {
		return ""
	}

	return strings.Replace(s, "*", "%", 2)
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	err := validate.Many(
		validate.Text("Search", opts.Search, 0, search.MaxQueryLen),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.Range("Omit", len(opts.Omit), 0, 50),
	)

	if opts.After.Key != "" {
		err = validate.Many(err, validate.LabelKey("After.Key", opts.After.Key))
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

func (opts renderData) QueryArgs() []sql.NamedArg {
	var afterServiceID string
	if opts.After.TargetType == assignment.TargetTypeService {
		afterServiceID = opts.After.TargetID
	}
	return []sql.NamedArg{
		sql.Named("keySearch", opts.KeySearch()),
		sql.Named("valueSearch", opts.ValueSearch()),
		sql.Named("afterKey", opts.After.Key),
		sql.Named("afterServiceID", afterServiceID),
		sql.Named("omit", pq.StringArray(opts.Omit)),
	}
}

func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]Label, error) {
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

	var result []Label
	var l Label
	var svcID sql.NullString
	for rows.Next() {
		err = rows.Scan(
			&l.Key,
			&l.Value,
			&svcID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "scan row")
		}

		switch {
		case svcID.Valid:
			l.Target = assignment.ServiceTarget(svcID.String)
		}

		result = append(result, l)
	}

	return result, nil
}
