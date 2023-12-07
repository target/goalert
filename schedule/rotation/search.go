package rotation

import (
	"context"
	"database/sql"
	"text/template"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SearchOptions allow filtering and paginating the list of rotations.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of rotation IDs to exclude from the results
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`

	// FavoritesOnly controls filtering the results to those marked as favorites by FavoritesUserID.
	FavoritesOnly bool `json:"b,omitempty"`

	// FavoritesFirst indicates that rotations marked as favorite (by FavoritesUserID) should be returned first (before any non-favorites).
	FavoritesFirst bool `json:"f,omitempty"`

	//FavoritesUserID is the userID associated with the rotation favorite
	FavoritesUserID string `json:"u,omitempty"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name       string `json:"n,omitempty"`
	IsFavorite bool   `json:"f,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(`
	SELECT
		rot.id, 
		rot.name, 
		rot.description, 
		rot.type, 
		rot.start_time, 
		rot.shift_length, 
		rot.time_zone, 
		fav IS DISTINCT FROM NULL
	FROM rotations rot
	{{if not .FavoritesOnly }}LEFT {{end}}JOIN user_favorites fav ON rot.id = fav.tgt_rotation_id AND {{if .FavoritesUserID}}fav.user_id = :favUserID{{else}}false{{end}}
	WHERE true
	{{if .Omit}}
		AND NOT rot.id = any(:omit)
	{{end}}
	{{if .Search}}
		AND ({{orderedPrefixSearch "search" "rot.name"}} OR {{contains "search" "rot.description"}} OR {{contains "search" "rot.name"}})
	{{end}}
	{{if .After.Name}}
		AND
		{{if not .FavoritesFirst}}
			lower(rot.name) > lower(:afterName)
		{{else if .After.IsFavorite}}
			((fav IS DISTINCT FROM NULL AND lower(rot.name) > lower(:afterName)) OR fav isnull)
		{{else}}
			(fav isnull AND lower(rot.name) > lower(:afterName))
		{{end}}
	{{end}}
	ORDER BY {{ .OrderBy }}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	if opts.FavoritesFirst {
		return "fav isnull, lower(rot.name)"
	}
	return "lower(rot.name)"
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
		err = validate.Many(err, validate.IDName("After.Name", opts.After.Name))
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
		sql.Named("search", opts.Search),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
		sql.Named("favUserID", opts.FavoritesUserID),
	}
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Rotation, error) {

	if opts == nil {
		opts = &SearchOptions{}
	}

	userCheck := permission.User
	if opts.FavoritesUserID != "" {
		userCheck = permission.MatchUser(opts.FavoritesUserID)
	}

	err := permission.LimitCheckAny(ctx, permission.System, userCheck)
	if err != nil {
		return nil, err
	}

	data, err := (*renderData)(opts).Normalize()
	if err != nil {
		return nil, err
	}
	query, args, err := search.RenderQuery(ctx, searchTemplate, data)
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

	var result []Rotation
	var r Rotation
	var tz string
	for rows.Next() {
		err = rows.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.Start, &r.ShiftLength, &tz, &r.isUserFavorite)
		if err != nil {
			return nil, err
		}
		loc, err := util.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
		r.Start = r.Start.In(loc)
		result = append(result, r)
	}

	return result, nil
}
