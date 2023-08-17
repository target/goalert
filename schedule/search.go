package schedule

import (
	"context"
	"database/sql"
	"text/template"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// SearchOptions allow filtering and paginating the list of schedules.
type SearchOptions struct {
	Search string `json:"s,omitempty"`

	// FavoritesUserID specifies the UserID whose favorite services want to be displayed.
	FavoritesUserID string `json:"u,omitempty"`

	// FavoritesOnly controls filtering the results to those marked as favorites by FavoritesUserID.
	FavoritesOnly bool `json:"g,omitempty"`

	// FavoritesFirst indicates that services marked as favorite (by FavoritesUserID) should be returned first (before any non-favorites).
	FavoritesFirst bool `json:"f,omitempty"`

	After SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of schedule IDs to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name       string `json:"n,omitempty"`
	IsFavorite bool   `json:"f"`
}

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(`
	SELECT
		sched.id,
		sched.name,
		sched.description,
		sched.time_zone,
		fav IS DISTINCT FROM NULL
	FROM schedules sched
	{{if not .FavoritesOnly }}
		LEFT {{end}}JOIN user_favorites fav ON sched.id = fav.tgt_schedule_id
			AND {{if .FavoritesUserID}}fav.user_id = :favUserID{{else}}false{{end}}
	WHERE true
	{{if .Omit}}
		AND NOT sched.id = any(:omit)
	{{end}}
	{{if .Search}}
		AND {{orderedSearch "search" "sched.name"}}
	{{end}}
	{{if .After.Name}}
		AND
		{{if not .FavoritesFirst}}
			lower(sched.name) > lower(:afterName)
		{{else if .After.IsFavorite}}
			((fav IS DISTINCT FROM NULL AND lower(sched.name) > lower(:afterName)) OR fav isnull)
		{{else}}
			(fav isnull AND lower(sched.name) > lower(:afterName))
		{{end}}
	{{end}}
	ORDER BY {{ .OrderBy }}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	if opts.FavoritesFirst {
		return "fav isnull, lower(sched.name)"
	}
	return "lower(sched.name)"
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

func (store *Store) Search(ctx context.Context, opts *SearchOptions) ([]Schedule, error) {
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

	rows, err := store.db.QueryContext(ctx, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Schedule
	var s Schedule
	var tz string
	for rows.Next() {
		err = rows.Scan(&s.ID, &s.Name, &s.Description, &tz, &s.isUserFavorite)
		if err != nil {
			return nil, err
		}
		loc, err := util.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
		s.TimeZone = loc
		result = append(result, s)
	}

	return result, nil
}
