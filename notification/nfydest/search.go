package nfydest

import "context"

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
func SearchByList[t any](items []t, opts SearchOptions, fn func(t) FieldValue) (*SearchResult, error) {
	return nil, nil
}
