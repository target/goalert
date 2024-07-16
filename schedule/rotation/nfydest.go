package rotation

import (
	"context"

	"github.com/google/uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/search"
	"github.com/target/goalert/validation"
)

const (
	DestTypeRotation = "builtin-rotation"
	FieldRotationID  = "rotation_id"

	FallbackIconURL = "builtin://rotation"
)

func DestFromID(rotationID uuid.UUID) gadb.DestV1 {
	return gadb.DestV1{
		Type: DestTypeRotation,
		Args: map[string]string{FieldRotationID: rotationID.String()},
	}
}

var (
	_ nfydest.Provider      = (*Store)(nil)
	_ nfydest.FieldSearcher = (*Store)(nil)
)

func (s *Store) ID() string { return DestTypeRotation }
func (s *Store) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	return &nfydest.TypeInfo{
		Type:                       DestTypeRotation,
		Name:                       "Rotation",
		Enabled:                    true,
		SupportsAlertNotifications: true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:        FieldRotationID,
			Label:          "Rotation",
			InputType:      "text",
			SupportsSearch: true,
		}},
	}, nil
}

func (s *Store) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	cfg := config.FromContext(ctx)

	r, err := s.FindRotation(ctx, args[FieldRotationID])
	if err != nil {
		return nil, err
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Rotation",
		LinkURL:     cfg.CallbackURL("/rotations/" + r.ID),
		Text:        r.Name,
	}, nil
}

func (s *Store) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldRotationID:
		_, err := s.FindRotation(ctx, value)
		return err
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *Store) FieldLabel(ctx context.Context, fieldID, value string) (string, error) {
	switch fieldID {
	case FieldRotationID:
		r, err := s.FindRotation(ctx, value)
		if err != nil {
			return "", err
		}
		return r.Name, nil
	}

	return "", validation.NewGenericError("unknown field ID")
}

func (r Rotation) AsField() nfydest.FieldValue {
	return nfydest.FieldValue{
		Value:      r.ID,
		Label:      r.Name,
		IsFavorite: r.isUserFavorite,
	}
}

func (r Rotation) Cursor() (string, error) {
	return search.Cursor(SearchCursor{
		Name:       r.Name,
		IsFavorite: r.isUserFavorite,
	})
}

func (so *SearchOptions) FromNotifyOptions(ctx context.Context, opts nfydest.SearchOptions) error {
	so.Search = opts.Search
	so.Omit = opts.Omit
	so.Limit = opts.Limit
	if opts.Cursor != "" {
		err := search.ParseCursor(opts.Cursor, &so.After)
		if err != nil {
			return err
		}
	}
	so.FavoritesFirst = true
	so.FavoritesUserID = permission.UserID(ctx)
	return nil
}

func (s *Store) SearchField(ctx context.Context, fieldID string, options nfydest.SearchOptions) (*nfydest.SearchResult, error) {
	switch fieldID {
	case FieldRotationID:
		return nfydest.SearchByCursorFunc(ctx, options, s.Search)
	}

	return nil, validation.NewGenericError("unknown field ID")
}
