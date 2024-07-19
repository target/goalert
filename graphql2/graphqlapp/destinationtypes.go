package graphqlapp

import (
	"context"
	"slices"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// builtin-types
const (
	destSMTP = "builtin-smtp-email"

	fieldEmailAddress = "email_address"
)

type (
	FieldValuePair         App
	DestinationDisplayInfo App
)

func (q *Query) DestinationFieldValueName(ctx context.Context, input graphql2.DestinationFieldValidateInput) (string, error) {
	return q.DestReg.FieldLabel(ctx, input.DestType, input.FieldID, input.Value)
}

func (q *Query) DestinationFieldSearch(ctx context.Context, input graphql2.DestinationFieldSearchInput) (*graphql2.FieldSearchConnection, error) {
	var opts nfydest.SearchOptions
	opts.Omit = input.Omit
	if input.First != nil {
		opts.Limit = *input.First
	}
	if input.After != nil {
		opts.Cursor = *input.After
	}
	if input.Search != nil {
		opts.Search = *input.Search
	}

	res, err := q.DestReg.SearchField(ctx, input.DestType, input.FieldID, opts)
	if err != nil {
		return nil, err
	}
	var nodes []graphql2.FieldSearchResult
	for _, v := range res.Values {
		nodes = append(nodes, graphql2.FieldSearchResult{
			FieldID:    input.FieldID,
			Value:      v.Value,
			Label:      v.Label,
			IsFavorite: v.IsFavorite,
		})
	}

	return &graphql2.FieldSearchConnection{
		Nodes: nodes,
		PageInfo: &graphql2.PageInfo{
			HasNextPage: res.HasNextPage,
			EndCursor:   &res.Cursor,
		},
	}, nil
}

func (q *Query) DestinationFieldValidate(ctx context.Context, input graphql2.DestinationFieldValidateInput) (bool, error) {
	switch input.DestType {
	case destSMTP:
		if input.FieldID != fieldEmailAddress {
			return false, validation.NewGenericError("unsupported field")
		}

		return validate.Email("Email", input.Value) == nil, nil
	}

	err := q.DestReg.ValidateField(ctx, input.DestType, input.FieldID, input.Value)
	if validation.IsClientError(err) {
		return false, nil
	}
	return err == nil, err
}

func (q *Query) DestinationTypes(ctx context.Context, isDynamicAction *bool) ([]nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	types := []nfydest.TypeInfo{
		{
			Type:                       destSMTP,
			Name:                       "Email",
			Enabled:                    cfg.SMTP.Enable,
			SupportsAlertNotifications: true,
			SupportsUserVerification:   true,
			SupportsStatusUpdates:      true,
			UserVerificationRequired:   true,
			RequiredFields: []nfydest.FieldConfig{{
				FieldID:            fieldEmailAddress,
				Label:              "Email Address",
				PlaceholderText:    "foobar@example.com",
				InputType:          "email",
				SupportsValidation: true,
			}},
			DynamicParams: []nfydest.DynamicParamConfig{{
				ParamID: "subject",
				Label:   "Subject",
				Hint:    "Subject of the email message.",
			}, {
				ParamID: "body",
				Label:   "Body",
				Hint:    "Body of the email message.",
			}},
		},
	}

	fromReg, err := q.DestReg.Types(ctx)
	if err != nil {
		return nil, err
	}
	types = append(types, fromReg...)

	slices.SortStableFunc(types, func(a, b nfydest.TypeInfo) int {
		if a.Enabled && !b.Enabled {
			return -1
		}
		if !a.Enabled && b.Enabled {
			return 1
		}

		// keep order for types that are both enabled or both disabled
		return 0
	})

	filtered := types[:0]
	for _, t := range types {
		if isDynamicAction != nil && *isDynamicAction != t.IsDynamicAction() {
			continue
		}

		filtered = append(filtered, t)
	}

	return filtered, nil
}
