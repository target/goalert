package schedule

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
	DestTypeSchedule = "builtin-schedule"
	FieldScheduleID  = "schedule_id"

	FallbackIconURL = "builtin://schedule"
)

var (
	_ nfydest.Provider      = (*Store)(nil)
	_ nfydest.FieldSearcher = (*Store)(nil)
)

func DestFromID(scheduleID uuid.UUID) gadb.DestV1 {
	return gadb.DestV1{
		Type: DestTypeSchedule,
		Args: map[string]string{FieldScheduleID: scheduleID.String()},
	}
}

func (s *Store) ID() string { return DestTypeSchedule }
func (s *Store) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	return &nfydest.TypeInfo{
		Type:                       DestTypeSchedule,
		Name:                       "Schedule",
		Enabled:                    true,
		SupportsAlertNotifications: true,
		RequiredFields: []nfydest.FieldConfig{{
			FieldID:        FieldScheduleID,
			Label:          "Schedule",
			InputType:      "text",
			SupportsSearch: true,
		}},
	}, nil
}

func (s *Store) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	cfg := config.FromContext(ctx)

	sched, err := s.FindOne(ctx, args[FieldScheduleID])
	if err != nil {
		return nil, err
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Schedule",
		LinkURL:     cfg.CallbackURL("/schedules/" + sched.ID),
		Text:        sched.Name,
	}, nil
}

func (s *Store) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldScheduleID:
		_, err := s.FindOne(ctx, value)
		return err
	}

	return validation.NewGenericError("unknown field ID")
}

func (s *Store) FieldLabel(ctx context.Context, fieldID, value string) (string, error) {
	switch fieldID {
	case FieldScheduleID:
		sched, err := s.FindOne(ctx, value)
		if err != nil {
			return "", err
		}
		return sched.Name, nil
	}

	return "", validation.NewGenericError("unknown field ID")
}

func (s Schedule) AsField() nfydest.FieldValue {
	return nfydest.FieldValue{
		Value:      s.ID,
		Label:      s.Name,
		IsFavorite: s.isUserFavorite,
	}
}

func (s Schedule) Cursor() (string, error) {
	return search.Cursor(SearchCursor{
		Name:       s.Name,
		IsFavorite: s.isUserFavorite,
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
	case FieldScheduleID:
		return nfydest.SearchByCursorFunc(ctx, options, s.Search)
	}

	return nil, validation.NewGenericError("unknown field ID")
}
