package user

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/user/contactmethod"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
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

	// CMValue is matched against the user's contact method phone number.
	CMValue string `json:"v,omitempty"`

	// CMType is matched against the user's contact method type.
	CMType contactmethod.Type `json:"t,omitempty"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name string `json:"n,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		usr.id, usr.name, usr.email, usr.role
	FROM users usr
	{{ if .CMValue }}
		JOIN user_contact_methods ucm ON ucm.user_id = usr.id
	{{ end }}
	WHERE true
	{{if .Omit}}
		AND not usr.id = any(:omit)
	{{end}}
	{{if .SearchStr}}
		AND usr.name ILIKE :search
	{{end}}
	{{if .After.Name}}
		AND lower(usr.name) > lower(:afterName)
	{{end}}
	{{ if .CMValue }}
		AND ucm.value = :CMValue
	{{ end }}
	{{ if .CMType }}
		AND ucm.type = :CMType
	{{ end }}
	ORDER BY lower(usr.name)
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
	if opts.CMValue != "" {
		err = validate.Many(err, validate.Phone("CMValue", opts.CMValue))
	}
	if opts.CMType != "" {
		if opts.CMValue == "" {
			err = validate.Many(err, validation.NewFieldError("CMValue", "is required"))
		}
		err = validate.Many(err, validate.OneOf("CMType", opts.CMType, contactmethod.TypeSMS, contactmethod.TypeVoice))
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
		sql.Named("CMValue", opts.CMValue),
		sql.Named("CMType", opts.CMType),
	}
}

func (db *DB) Search(ctx context.Context, opts *SearchOptions) ([]User, error) {
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

	var result []User
	var u User
	for rows.Next() {
		err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}
