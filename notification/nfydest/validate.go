package nfydest

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation"
)

type DestValidator interface {
	ValidateDest(ctx context.Context, dest gadb.DestV1) error
}

// DestArgError is returned when a destination argument is invalid.
type DestArgError struct {
	FieldID string
	Err     error
}

func (e *DestArgError) Error() string { return fmt.Sprintf("field %s: %s", e.FieldID, e.Err) }

func (e *DestArgError) ClientError() bool { return true }

func (r *Registry) ValidateDest(ctx context.Context, dest gadb.DestV1) error {
	p := r.Provider(dest.Type)
	if p == nil {
		return ErrUnknownType
	}

	info, err := p.TypeInfo(ctx)
	if err != nil {
		return err
	}

	if !info.Enabled {
		return ErrNotEnabled
	}

	if dest.Args == nil {
		dest.Args = make(map[string]string)
	}

	if v, ok := p.(DestValidator); ok {
		// Some providers may need/want to validate all args at once.
		err := v.ValidateDest(ctx, dest)
		// If we get `ErrUnsupported`, we'll fall back to the field-by-field validation, this can happen if the provider implements the interface, but the backing implementation (e.g., external plugin) doesn't support it.
		if !errors.Is(err, ErrUnsupported) {
			// If the provider implements the DestValidator interface, we'll use that for validation. We return err, even if it's nil UNLESS it's ErrUnsupported.
			return err
		}
	}

	fieldNames := make([]string, 0, len(info.RequiredFields))
	for _, f := range info.RequiredFields {
		fieldNames = append(fieldNames, f.FieldID)
	}

	// Make sure we reject any fields that are not expected.
	for fName := range dest.Args {
		if slices.Contains(fieldNames, fName) {
			continue
		}

		return &DestArgError{
			FieldID: fName,
			Err:     fmt.Errorf("unexpected field"),
		}
	}

	// Make sure all required fields are valid, which may be allowed to be empty (thus we don't iterate over dest.Args).
	for _, f := range info.RequiredFields {
		err := p.ValidateField(ctx, f.FieldID, dest.Args[f.FieldID])
		if errors.Is(err, sql.ErrNoRows) {
			err = validation.NewGenericError("does not exist")
		}
		if validation.IsClientError(err) {
			return &DestArgError{
				FieldID: f.FieldID,
				Err:     err,
			}
		}
		if err != nil {
			return fmt.Errorf("validate field %s: %w", f.FieldID, err)
		}
	}

	// Since we have no extra/unknown fields, and all required fields are valid, we've validated the destination.
	return nil
}
