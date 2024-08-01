package user

import (
	"context"
	"database/sql"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions allow filtering and paginating the list of users.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of user IDs to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`

	DestArgs map[string]string `json:"da,omitempty"`
	DestType string            `json:"dt,omitempty"`

	// FavoritesUserID specifies the UserID whose favorite users want to be displayed.
	FavoritesUserID string `json:"u,omitempty"`

	// FavoritesOnly controls filtering the results to those marked as favorites by FavoritesUserID.
	FavoritesOnly bool `json:"g,omitempty"`

	// FavoritesFirst indicates the user marked as favorite (by FavoritesUserID) should be returned first (before any non-favorites).
	FavoritesFirst bool `json:"f,omitempty"`
}

func (so *SearchOptions) FromNotifyOptions(ctx context.Context, opts nfydest.SearchOptions) error {
	so.Search = opts.Search
	so.Omit = opts.Omit
	so.Limit = opts.Limit
	if opts.Cursor != "" {
		err := search.ParseCursor(opts.Cursor, &so.After)
		if err != nil {
			return err
		}
	}
	so.FavoritesFirst = true
	so.FavoritesUserID = permission.UserID(ctx)
	return nil
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name       string `json:"n,omitempty"`
	IsFavorite bool   `json:"f,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(`
	SELECT DISTINCT ON ({{ .OrderBy }})
		usr.id, usr.name, usr.email, usr.role, fav IS DISTINCT FROM NULL
	FROM users usr
	{{ if eq (len .DestArgs) 0 }}
		JOIN user_contact_methods ucm ON ucm.user_id = usr.id
	{{ end }}
	{{if not .FavoritesOnly}}
		LEFT {{end}} JOIN user_favorites fav on usr.id = fav.tgt_user_id 
			AND {{if .FavoritesUserID}} fav.user_id = :favUserID{{else}}false{{end}}
	WHERE true
	{{if .Omit}}
		AND not usr.id = any(:omit)
	{{end}}
	{{if .Search}}
		AND ({{orderedPrefixSearch "search" "usr.name"}}  OR {{contains "search" "usr.name"}})
	{{end}}
	{{if .After.Name}}
		AND {{if not .FavoritesFirst}}
			lower(usr.name) > lower(:afterName)
		{{else if .After.IsFavorite}}
			((fav IS DISTINCT FROM NULL AND lower(usr.name) > lower(:afterName)) OR fav isnull)
		{{else}}
			(fav isnull AND lower(usr.name) > lower(:afterName))
		{{end}}
	{{end}}
	{{ if .Email }}
		AND lower(usr.email) = lower(:Email)
	{{ end }}
	{{ if eq (len .DestArgs) 0 }}
		AND ucm.dest->'Args' = :DestArgs
	{{ end }}
	{{ if .DestType }}
		AND ucm.dest->>'Type' = :DestType
	{{ end }}
	ORDER BY {{ .OrderBy }}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	if opts.FavoritesFirst {
		return "fav isnull, lower(usr.name), usr.id"
	}
	return "lower(usr.name), usr.id"
}

func (opts renderData) SearchString() string {
	if opts.Email() != "" {
		return ""
	}

	return opts.Search
}

func (opts renderData) Email() string {
	if !strings.HasPrefix(opts.Search, "email=") {
		return ""
	}
	return strings.TrimPrefix(opts.Search, "email=")
}

func (opts renderData) Normalize() (*renderData, error) {
	if opts.Limit == 0 {
		opts.Limit = search.DefaultMaxResults
	}

	if opts.DestType != "" && opts.DestArgs == nil {
		return nil, validation.NewGenericError("DestArgs must be set when DestType is set")
	}

	err := validate.Many(
		validate.Search("Search", opts.Search),
		validate.Range("Limit", opts.Limit, 0, search.MaxResults),
		validate.ManyUUID("Omit", opts.Omit, 50),
	)
	if opts.After.Name != "" {
		err = validate.Many(err, validate.Name("After.Name", opts.After.Name))
	}
	if opts.FavoritesOnly || opts.FavoritesFirst || opts.FavoritesUserID != "" {
		err = validate.Many(err, validate.UUID("FavoritesUserID", opts.FavoritesUserID))
	}
	if err != nil {
		return nil, err
	}

	return &opts, err
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("search", opts.SearchString()),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
		sql.Named("DestArgs", opts.DestArgs),
		sql.Named("DestType", opts.DestType),
		sql.Named("favUserID", opts.FavoritesUserID),
		sql.Named("Email", opts.Email()),
	}
}

func (u User) Cursor() (string, error) {
	return search.Cursor(SearchCursor{
		Name:       u.Name,
		IsFavorite: u.isUserFavorite,
	})
}

func (u User) AsField() nfydest.FieldValue {
	return nfydest.FieldValue{
		Value:      u.ID,
		Label:      u.Name,
		IsFavorite: u.isUserFavorite,
	}
}

// Search performs a paginated search of users with the given options.
func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]User, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	userCheck := permission.User
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
	if opts.FavoritesUserID != "" {
		userCheck = permission.MatchUser(opts.FavoritesUserID)
	}
	err = permission.LimitCheckAny(ctx, permission.System, userCheck)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []User
	var u User
	for rows.Next() {
		err = rows.Scan(&u.ID, &u.Name, &u.Email, &u.Role, &u.isUserFavorite)
		if err != nil {
			return nil, err
		}
		result = append(result, u)
	}

	return result, nil
}
