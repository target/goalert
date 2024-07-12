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

func (DisplayInfo) IsInlineDisplayInfo() {}
