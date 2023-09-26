package timezone

import (
	"context"
	"slices"
	"strings"

	"github.com/target/goalert/permission"
)

// SearchOptions allow filtering and paginating the list of timezones.
type SearchOptions struct {
	Search string       `json:"s,omitempty"`
	After  SearchCursor `json:"a,omitempty"`

	// Omit specifies a list of timezone names to exclude from the results.
	Omit []string `json:"o,omitempty"`

	Limit int `json:"-"`
}

// SearchCursor is used to indicate a position in a paginated list.
type SearchCursor struct {
	Name string `json:"n,omitempty"`
}

func (store *Store) Search(ctx context.Context, opts *SearchOptions) ([]string, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		opts = &SearchOptions{}
	}

	results := make([]string, 0, len(zones))
	for _, zone := range zones {
		if opts.Search != "" && !strings.Contains(strings.ToLower(zone), strings.ToLower(opts.Search)) {
			continue
		}
		if slices.Contains(opts.Omit, zone) {
			continue
		}
		if opts.After.Name != "" && zone <= opts.After.Name {
			continue
		}

		results = append(results, zone)

		if len(results) >= opts.Limit {
			break
		}
	}

	return results, nil
}
