package nfydest

import "context"

type Provider interface {
	ID() string
	TypeInfo(ctx context.Context) (*TypeInfo, error)

	ValidateField(ctx context.Context, fieldID, value string) (ok bool, err error)
	DisplayInfo(ctx context.Context, args map[string]string) (*DisplayInfo, error)
}

type DisplayInfo struct {
	Text        string
	IconURL     string
	IconAltText string
	LinkURL     string
}

type SearchOptions struct {
	Search string
	Omit   []string
	Cursor string
	Limit  int
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

type FieldSearcher interface {
	SearchField(ctx context.Context, fieldID string, options SearchOptions) (*SearchResult, error)
	FieldLabel(ctx context.Context, fieldID, value string) (string, error)
}
