package alert

import (
	"context"

	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

const (
	DestTypeAlert = "builtin-alert"

	ParamSummary = "summary"
	ParamDetails = "details"
	ParamDedup   = "dedup"
	ParamClose   = "close"

	FallbackIconURL = "builtin://alert"
)

var _ nfydest.Provider = (*Store)(nil)

func (s *Store) ID() string { return DestTypeAlert }
func (s *Store) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	return &nfydest.TypeInfo{
		Type:            DestTypeAlert,
		Name:            "Alert",
		Enabled:         true,
		SupportsSignals: true,
		DynamicParams: []nfydest.DynamicParamConfig{{
			ParamID: ParamSummary,
			Label:   "Summary",
			Hint:    "Short summary of the alert (used for things like SMS).",
		}, {
			ParamID: ParamDetails,
			Label:   "Details",
			Hint:    "Full body (markdown) text of the alert.",
		}, {
			ParamID: ParamDedup,
			Label:   "Dedup",
			Hint:    "Stable identifier for de-duplication and closing existing alerts.",
		}, {
			ParamID: ParamClose,
			Label:   "Close",
			Hint:    "If true, close an existing alert.",
		}},
	}, nil
}

func (s *Store) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURL,
		IconAltText: "Alert",
		Text:        "Create new alert",
	}, nil
}

func (s *Store) ValidateField(ctx context.Context, fieldID, value string) error {
	// Alert destination has no valid required fields.
	return validation.NewGenericError("unknown field ID")
}
