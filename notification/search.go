package notification

import (
	"context"
	"database/sql"
	"text/template"

	"github.com/pkg/errors"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
)

// SearchOptions allow filtering and paginating the list of rotations.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of rotation IDs to exclude from the results
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	SrcValue string `json:"n,omitempty"`
}

const _ = `
func (a *Query) DebugMessages(ctx context.Context, input *graphql2.DebugMessagesInput) ([]graphql2.DebugMessage, error) {
	if input.CreatedAfter != nil {
		db = db.Where("created_at >= ?", *input.CreatedAfter)
	}
	if input.CreatedBefore != nil {
		db = db.Where("created_at < ?", *input.CreatedBefore)
	}
	if input.First != nil {
		err = validate.Range("first", *input.First, 0, 1000)
		if err != nil {
			return nil, err
		}
		db = db.Limit(*input.First)
	} else {
		db = db.Limit(search.DefaultMaxResults)
	}
	db.
		Preload("User", sqlutil.Columns("ID", "Name")).
		Preload("Service", sqlutil.Columns("ID", "Name")).
		Preload("Channel", sqlutil.Columns("ID", "Type", "Value")).
		Preload("ContactMethod", sqlutil.Columns("ID", "Type", "Value")).
		Order("created_at DESC").
		Find(&msgs).Error

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
	AND {{prefixSearch "search" "rot.name"}}
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
`

// todo
var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(``))

type renderData SearchOptions

// todo
func (opts renderData) Normalize() (*renderData, error) {
	return nil, nil
}

// todo
func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{}
}

func (s *Store) Search(ctx context.Context, opts *SearchOptions) ([]MessageLog, error) {
	if opts == nil {
		opts = &SearchOptions{}
	}

	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
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

	var result []MessageLog
	var l MessageLog
	for rows.Next() {
		err = rows.Scan(
			&l.ID,
			&l.CreatedAt,
			&l.LastStatus,
			&l.MessageType,
			&l.LastStatus,
			&l.StatusDetails,
			&l.SrcValue,
			&l.AlertID,
			&l.ProviderMsgID,
			&l.User.ID,
			&l.ContactMethod.ID,
			&l.Channel.ID,
			&l.Service.ID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, l)
	}

	return result, nil
}
