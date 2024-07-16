package user

import (
	"context"

	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

const (
	DestTypeUser = "builtin-user"
	FieldUserID  = "user_id"

	FallbackIconURL = "builtin://user"
)

var (
	_ nfydest.Provider      = (*Store)(nil)
	_ nfydest.FieldSearcher = (*Store)(nil)
)

func DestFromID(userID string) gadb.DestV1 {
	return gadb.DestV1{
		Type: DestTypeUser,
		Args: map[string]string{FieldUserID: userID},
	}
}

func (s *Store) ID() string { return DestTypeUser }
func (s *Store) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	return &nfydest.TypeInfo{
		Type:                       DestTypeUser,
		Name:                       "User",
		Enabled:                    true,
		SupportsAlertNotifications: true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:        FieldUserID,
			Label:          "User",
			InputType:      "text",
			SupportsSearch: true,
		}},
	}, nil
}

func (s *Store) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	cfg := config.FromContext(ctx)

	u, err := s.FindOne(ctx, args[FieldUserID])
	if err != nil {
		return nil, err
	}

	return &nfydest.DisplayInfo{
		IconURL:     cfg.CallbackURL("/api/v2/user-avatar/" + u.ID),
		IconAltText: "User",
		LinkURL:     cfg.CallbackURL("/users/" + u.ID),
		Text:        u.Name,
	}, nil
}

func (s *Store) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldUserID:
		_, err := s.FindOne(ctx, value)
		return err
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *Store) FieldLabel(ctx context.Context, fieldID, value string) (string, error) {
	switch fieldID {
	case FieldUserID:
		u, err := s.FindOne(ctx, value)
		if err != nil {
			return "", err
		}
		return u.Name, nil
	}

	return "", validation.NewGenericError("unknown field ID")
}

func (s *Store) SearchField(ctx context.Context, fieldID string, options nfydest.SearchOptions) (*nfydest.SearchResult, error) {
	switch fieldID {
	case FieldUserID:
		return nfydest.SearchByCursorFunc(ctx, options, s.Search)
	}

	return nil, validation.NewGenericError("unknown field ID")
}
