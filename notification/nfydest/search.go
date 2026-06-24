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

type Fieldable interface {
	AsField() FieldValue
}
type Cursorable interface {
	Cursor() (string, error)
	Fieldable
}

type OptionFrommer interface {
	FromNotifyOptions(context.Context, SearchOptions) error
}

func SearchByCursorFunc[OptionType any, POptionType interface {
	*OptionType
	OptionFrommer
}, Result Cursorable](ctx context.Context, opts SearchOptions, searchFn func(context.Context, *OptionType) ([]Result, error)) (*SearchResult, error) {
	if opts.Limit <= 0 {
		opts.Limit = 15
	}
	origLimit := opts.Limit
	opts.Limit++ // Fetch one more to determine if there is a next page.

	var searchOpts OptionType
	p := POptionType(&searchOpts)
	err := p.FromNotifyOptions(ctx, opts)
	if err != nil {
		return nil, err
	}

	results, err := searchFn(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}

	var res SearchResult
	if len(results) > origLimit {
		res.HasNextPage = true
		res.Cursor, err = results[origLimit].Cursor()
		if err != nil {
			return nil, err
		}
		results = results[:origLimit]
	}
	for _, r := range results {
		res.Values = append(res.Values, r.AsField())
	}

	return &res, nil
}

// SearchByListFunc allows returning a SearchResult from a function that returns a list of Fieldable items.
func SearchByListFunc[T Fieldable](ctx context.Context, searchOpts SearchOptions, listFn func(context.Context) ([]T, error)) (*SearchResult, error) {
	if searchOpts.Limit <= 0 {
		searchOpts.Limit = 15 // Default limit.
	}

	_items, err := listFn(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]FieldValue, 0, len(_items))
	for _, item := range _items {
		items = append(items, item.AsField())
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
