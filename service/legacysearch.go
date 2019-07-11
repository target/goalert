package service

import (
	"bytes"
	"context"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation/validate"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

// LegacySearchOptions contains criteria for filtering and sorting services.
type LegacySearchOptions struct {
	// Search is matched case-insensitive against the service name and description.
	Search string

	// FavoritesUserID specifies the UserID whose favorite services want to be displayed.
	FavoritesUserID string

	// FavoritesOnly controls filtering the results to those marked as favorites by FavoritesUserID.
	FavoritesOnly bool

	// FavoritesFirst indicates that services marked as favorite (by FavoritesUserID) should be returned first (before any non-favorites).
	FavoritesFirst bool

	// Limit, if not zero, will limit the number of results.
	Limit int
}

var legacySearchTemplate = template.Must(template.New("search").Parse(`
	SELECT DISTINCT ON ({{ .OrderBy }})
		svc.id,
		svc.name,
		svc.description,
		ep.name,
		fav notnull
	FROM services svc
	JOIN escalation_policies ep ON ep.id = svc.escalation_policy_id
	{{if not .FavoritesOnly }}LEFT {{end}}JOIN user_favorites fav ON svc.id = fav.tgt_service_id AND fav.user_id = $1
	{{if and .LabelKey .LabelNegate}}
		WHERE svc.id NOT IN (
			SELECT tgt_service_id
			FROM labels
			WHERE
				tgt_service_id NOTNULL AND
				key = $2
				{{if ne .LabelValue "*"}} AND value = $3{{end}}
		)
	{{else if .LabelKey}}
		JOIN labels l ON
			l.tgt_service_id = svc.id AND
			l.key = $2
			{{if ne .LabelValue "*"}} AND value = $3{{end}}
	{{else}}
		WHERE $2 = '' OR svc.name ILIKE $2 OR svc.description ILIKE $2
	{{end}}
	ORDER BY {{ .OrderBy }}
	{{if ne .Limit 0}}LIMIT {{.Limit}}{{end}}
`))

// LegacySearch will return a list of matching services and the total number of matches available.
func (db *DB) LegacySearch(ctx context.Context, opts *LegacySearchOptions) ([]Service, error) {
	if opts == nil {
		opts = &LegacySearchOptions{}
	}

	userCheck := permission.User
	if opts.FavoritesUserID != "" {
		userCheck = permission.MatchUser(opts.FavoritesUserID)
	}

	err := permission.LimitCheckAny(ctx, permission.System, userCheck)
	if err != nil {
		return nil, err
	}

	err = validate.Search("Search", opts.Search)
	if opts.FavoritesOnly || opts.FavoritesFirst || opts.FavoritesUserID != "" {
		err = validate.Many(err, validate.UUID("FavoritesUserID", opts.FavoritesUserID))
	}
	if err != nil {
		return nil, err
	}

	var renderContext struct {
		LegacySearchOptions

		OrderBy string

		LabelKey    string
		LabelValue  string
		LabelNegate bool

		Limit int
	}
	renderContext.LegacySearchOptions = *opts

	var parts []string
	if opts.FavoritesFirst {
		parts = append(parts, "fav")
	}
	parts = append(parts,
		"lower(svc.name)", // use lower because we already have a unique index that does this
		"svc.name",
	)
	renderContext.OrderBy = strings.Join(parts, ",")

	queryArgs := []interface{}{
		opts.FavoritesUserID,
	}

	// case sensitive searching for labels
	if idx := strings.Index(opts.Search, "="); idx > -1 {
		renderContext.LabelKey = opts.Search[:idx]
		if strings.HasSuffix(renderContext.LabelKey, "!") {
			renderContext.LabelNegate = true
			renderContext.LabelKey = strings.TrimSuffix(renderContext.LabelKey, "!")
		}
		renderContext.LabelValue = opts.Search[idx+1:]
		if renderContext.LabelValue == "" {
			renderContext.LabelValue = "*"
		}
		// skip validating LabelValue if search wildcard character or < 3 characters
		if renderContext.LabelValue == "*" || len(renderContext.LabelValue) < 3 {
			err = validate.LabelKey("LabelKey", renderContext.LabelKey)
		} else {
			err = validate.Many(
				validate.LabelKey("LabelKey", renderContext.LabelKey),
				validate.LabelValue("LabelValue", renderContext.LabelValue),
			)
		}
		if err != nil {
			return nil, err
		}

		queryArgs = append(queryArgs, renderContext.LabelKey)
		if renderContext.LabelValue != "*" {
			queryArgs = append(queryArgs, renderContext.LabelValue)
		}

	} else {
		opts.Search = "%" + search.Escape(opts.Search) + "%"
		queryArgs = append(queryArgs, opts.Search)
	}

	buf := new(bytes.Buffer)
	err = legacySearchTemplate.Execute(buf, renderContext)
	if err != nil {
		return nil, errors.Wrap(err, "render query")
	}

	rows, err := db.db.QueryContext(ctx, buf.String(), queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Service
	for rows.Next() {
		var s Service
		err = rows.Scan(&s.ID, &s.Name, &s.Description, &s.epName, &s.isUserFavorite)
		if err != nil {
			return nil, err
		}

		result = append(result, s)
	}

	return result, nil
}
