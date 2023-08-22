package service

import (
	"context"
	"database/sql"
	"strings"
	"text/template"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

// SearchOptions contains criteria for filtering and sorting services.
type SearchOptions struct {
	// Search is matched case-insensitive against the service name and description.
	Search string `json:"s,omitempty"`

	// FavoritesUserID specifies the UserID whose favorite services want to be displayed.
	FavoritesUserID string `json:"u,omitempty"`

	// FavoritesOnly controls filtering the results to those marked as favorites by FavoritesUserID.
	FavoritesOnly bool `json:"o,omitempty"`

	// Omit specifies a list of service IDs to exclude from the results.
	Omit []string `json:"m,omitempty"`

	// FavoritesFirst indicates that services marked as favorite (by FavoritesUserID) should be returned first (before any non-favorites).
	FavoritesFirst bool `json:"f,omitempty"`

	// Limit will limit the number of results.
	Limit int `json:"-"`

	After SearchCursor `json:"a,omitempty"`
}

type SearchCursor struct {
	Name       string `json:"n"`
	IsFavorite bool   `json:"f"`
}

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(`
	SELECT{{if .LabelKey}} DISTINCT ON ({{ .OrderBy }}){{end}}
		svc.id,
		svc.name,
		svc.description,
		svc.escalation_policy_id,
		fav IS DISTINCT FROM NULL
	FROM services svc
	{{if not .FavoritesOnly }}LEFT {{end}}JOIN user_favorites fav ON svc.id = fav.tgt_service_id AND {{if .FavoritesUserID}}fav.user_id = :favUserID{{else}}false{{end}}
	{{if and .IntegrationKey}}
		JOIN integration_keys intKey ON
			intKey.service_id = svc.id AND
			intKey.id = :integrationKey
	{{end}}
	{{if and .LabelKey (not .LabelNegate)}}
		JOIN labels l ON
			l.tgt_service_id = svc.id AND
			l.key = :labelKey
			{{if ne .LabelValue "*"}} AND value = :labelValue{{end}}
	{{end}}
	WHERE true
	{{if .Omit}}
		AND not svc.id = any(:omit)
	{{end}}
	{{- if and .LabelKey .LabelNegate}}
		AND svc.id NOT IN (
			SELECT tgt_service_id
			FROM labels
			WHERE
				tgt_service_id NOTNULL AND
				key = :labelKey
				{{if ne .LabelValue "*"}} AND value = :labelValue{{end}}
		)
	{{end}}
	{{- if and .Search (not .LabelKey) (not .IntegrationKey)}}
		AND {{orderedPrefixSearch "search" "svc.name"}}
	{{- end}}
	{{- if .After.Name}}
		AND
		{{if not .FavoritesFirst}}
			lower(svc.name) > lower(:afterName)
		{{else if .After.IsFavorite}}
			((fav IS DISTINCT FROM NULL AND lower(svc.name) > lower(:afterName)) OR fav isnull)
		{{else}}
			(fav isnull AND lower(svc.name) > lower(:afterName))
		{{end}}
	{{- end}}
	ORDER BY {{ .OrderBy }}
	LIMIT {{.Limit}}
`))

type renderData SearchOptions

func (opts renderData) OrderBy() string {
	if opts.FavoritesFirst {
		return "fav isnull, lower(svc.name)"
	}

	return "lower(svc.name)"
}

func (opts renderData) IntegrationKey() string {
	if !strings.Contains(opts.Search, "token=") {
		return ""
	}
	return opts.Search[6:42]
}

func (opts renderData) LabelKey() string {
	searchStr := opts.Search
	if strings.Contains(opts.Search, "token=") {
		// strip token string
		searchStr = opts.Search[42:]
		searchStr = strings.TrimSpace(searchStr)
	}
	idx := strings.IndexByte(searchStr, '=')
	if idx == -1 {
		return ""
	}
	return strings.TrimSuffix(searchStr[:idx], "!") // if `!=`` is used
}
func (opts renderData) LabelValue() string {
	searchStr := opts.Search
	if strings.Contains(opts.Search, "token=") {
		// strip token string
		searchStr = opts.Search[42:]
		searchStr = strings.TrimSpace(searchStr)
	}
	idx := strings.IndexByte(searchStr, '=')
	if idx == -1 {
		return ""
	}
	val := searchStr[idx+1:]
	if val == "" {
		return "*"
	}
	return val
}
func (opts renderData) LabelNegate() bool {
	idx := strings.IndexByte(opts.Search, '=')
	if idx < 1 {
		return false
	}

	return opts.Search[idx-1] == '!'
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
	if opts.IntegrationKey() != "" {
		err = validate.Search("IntegrationKey", opts.IntegrationKey())
	}
	if opts.LabelKey() != "" {
		err = validate.Search("LabelKey", opts.LabelKey())
		if opts.LabelValue() != "*" {
			err = validate.Many(err,
				validate.LabelValue("LabelValue", opts.LabelValue()),
			)
		}
	}
	if err != nil {
		return nil, err
	}

	return &opts, nil
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{
		sql.Named("favUserID", opts.FavoritesUserID),
		sql.Named("integrationKey", opts.IntegrationKey()),
		sql.Named("labelKey", opts.LabelKey()),
		sql.Named("labelValue", opts.LabelValue()),
		sql.Named("labelNegate", opts.LabelNegate()),
		sql.Named("search", opts.Search),
		sql.Named("afterName", opts.After.Name),
		sql.Named("omit", sqlutil.UUIDArray(opts.Omit)),
	}
}

// Search will return a list of matching services and the total number of matches available.
func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]Service, error) {
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
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Service
	for rows.Next() {
		var s Service
		err = rows.Scan(&s.ID, &s.Name, &s.Description, &s.EscalationPolicyID, &s.isUserFavorite)
		if err != nil {
			return nil, err
		}

		result = append(result, s)
	}

	return result, nil
}
