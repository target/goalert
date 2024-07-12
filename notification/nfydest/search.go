package nfydest

import (
	"context"
	"slices"
	"sort"
	"strings"
)

type FieldSearcher interface {
	SearchField(ctx context.Context, fieldID string, options SearchOptions) (*SearchResult, error)
	FieldLabel(ctx context.Context, fieldID, value string) (string, error)
}

type SearchResult struct {
	HasNextPage bool
	Cursor      string
	Values      []FieldValue
}

type FieldValue struct {
	Value      string
	Label      string
	IsFavorite bool
}

type SearchOptions struct {
	Search string
	Omit   []string
	Cursor string
	Limit  int
}

// SearchByList allows returning a SearchResult from a list of items, handling pagination and filtering.
func SearchByList[t any](_items []t, searchOpts SearchOptions, fn func(t) FieldValue) (*SearchResult, error) {
	if searchOpts.Limit <= 0 {
		searchOpts.Limit = 15 // Default limit.
	}

	items := make([]FieldValue, len(_items))
	for i, item := range _items {
		items[i] = fn(item)
	}
	// Sort by name, case-insensitive, then sensitive.
	sort.Slice(items, func(i, j int) bool {
		iLabel, jLabel := strings.ToLower(items[i].Label), strings.ToLower(items[j].Label)

		if iLabel != jLabel {
			return iLabel < jLabel
		}
		return items[i].Label < items[j].Label
	})

	// No DB search, so we manually filter for the cursor and search strings.
	searchOpts.Search = strings.ToLower(searchOpts.Search)
	filtered := items[:0]
	for _, item := range items {
		lowerName := strings.ToLower(item.Label)
		if !strings.Contains(lowerName, searchOpts.Search) {
			continue
		}
		if searchOpts.Cursor != "" && item.Label <= searchOpts.Cursor {
			continue
		}
		if slices.Contains(searchOpts.Omit, item.Value) {
			continue
		}
		filtered = append(filtered, item)
	}
	items = filtered

	hasNextPage := len(items) > searchOpts.Limit
	if hasNextPage {
		items = items[:searchOpts.Limit]
	}
	var cursor string
	if len(items) > 0 {
		cursor = items[len(items)-1].Label
	}

	return &SearchResult{
		HasNextPage: hasNextPage,
		Cursor:      cursor,
		Values:      items,
	}, nil
}
