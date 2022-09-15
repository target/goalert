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

var searchTemplate = template.Must(template.New("search").Funcs(search.Helpers()).Parse(``))

type renderData SearchOptions

func (opts renderData) Normalize() (*renderData, error) {
	return nil, nil
}

func (opts renderData) QueryArgs() []sql.NamedArg {
	return []sql.NamedArg{}
}

func (s *Store) FindOneMessageLog(ctx context.Context, id string) (*MessageLog, error) {
	return &MessageLog{}, nil
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
			&l.UserID,
			&l.ContactMethodID,
			&l.ChannelID,
			&l.ServiceID,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, l)
	}

	return result, nil
}
