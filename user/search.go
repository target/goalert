package user

import (
	"context"
	"database/sql"
	"github.com/target/goalert/util/sqlutil"
	"text/template"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"

	"github.com/target/goalert/user/contactmethod"
)

// SearchOptions allow filtering and paginating the list of users.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of user IDs to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
	//CmValue is matched against the user's contact method phone number.
	CmValue string `json:"v,omitempty"`
	//CmType is matched against the user's contact method type.
	CmType contactmethod.Type `json:"t,omitempty"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name string `json:"n,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Parse(`
	SELECT
		usr.id, usr.name, usr.email, usr.role
	FROM users usr
	{{ if .CmValue }}
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
	{{ if .CmValue }}
		AND ucm.value = :cmValue
	{{ end }}
	{{ if .CmType }}
		AND ucm.type = :cmType
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
	if opts.CmValue != "" {
		err = validate.Phone("CmValue", opts.CmValue)
	}
	if opts.CmType != "" {
		if opts.CmValue == "" {
			err = validation.NewFieldError("CmValue", "must be provided")
		} else {
			err = validate.OneOf("CmType", opts.CmType, contactmethod.TypeSMS, contactmethod.TypeVoice)
		}
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("search", opts.SearchStr()),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
		sql.Named("cmValue", opts.CmValue),
		sql.Named("cmType", opts.CmType),
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
