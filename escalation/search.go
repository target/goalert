package escalation

import (
	"context"
	"database/sql"
	"text/template"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SearchOptions allow filtering and paginating the list of escalation policies.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// FavoritesUserID specifies the UserID whose favorite escalation policies want to be displayed.
	FavoritesUserID string `json:"u,omitempty"`

	// FavoritesOnly controls filtering the results to those marked as favorites by FavoritesUserID.
	FavoritesOnly bool `json:"g,omitempty"`

	// FavoritesFirst indicates that escalation policy marked as favorite (by FavoritesUserID) should be returned first (before any non-favorites).
	FavoritesFirst bool `json:"f,omitempty"`

	// Omit specifies a list of policy IDs to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name       string `json:"n,omitempty"`
	IsFavorite bool   `json:"f,omitempty"`
}

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(`
	SELECT
		pol.id,
		pol.name,
		pol.description,
		pol.repeat,
		fav IS DISTINCT FROM NULL
	FROM escalation_policies pol
	{{if not .FavoritesOnly }}
		LEFT {{end}}JOIN user_favorites fav ON pol.id = fav.tgt_escalation_policy_id
			AND {{if .FavoritesUserID}}fav.user_id = :favUserID{{else}}false{{end}}
	WHERE true
	{{if .Omit}}
		AND NOT pol.id = any(:omit)
	{{end}}
	{{if .Search}}
		AND {{orderedPrefixSearch "search" "pol.name"}}
	{{end}}
	{{if .After.Name}}
		AND
		{{if not .FavoritesFirst}}
			lower(pol.name) > lower(:afterName)
		{{else if .After.IsFavorite}}
			((fav IS DISTINCT FROM NULL AND lower(pol.name) > lower(:afterName)) OR fav isnull)
		{{else}}
			(fav isnull AND lower(pol.name) > lower(:afterName))
		{{end}}
	{{end}}
	ORDER BY {{ .OrderBy }}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	if opts.FavoritesFirst {
		return "fav isnull, lower(pol.name)"
	}
	return "lower(pol.name)"
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

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &SearchOptions{}
	}
	userCheck := permission.User
	if opts.FavoritesUserID != "" {
		userCheck = permission.MatchUser(opts.FavoritesUserID)
	}
	err = permission.LimitCheckAny(ctx, permission.System, userCheck)
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

	var result []Policy
	var p Policy
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat, &p.isUserFavorite)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}
