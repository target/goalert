package nfydest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation"
)

type ParamValidator interface {
	ValidateParam(ctx context.Context, paramID, value string) error
}

type ActionValidator interface {
	ValidateAction(ctx context.Context, act gadb.UIKActionV1) error
}

type ActionParamError struct {
	ParamID string
	Err     error
}

func (e *ActionParamError) Error() string { return fmt.Sprintf("parameter %s: %s", e.ParamID, e.Err) }

func (r *Registry) ValidateAction(ctx context.Context, act gadb.UIKActionV1) error {
	p := r.Provider(act.Dest.Type)
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

	err = r.ValidateDest(ctx, act.Dest)
	if err != nil {
		return err
	}

	if act.Params == nil {
		act.Params = make(map[string]string)
	}

	if v, ok := p.(ActionValidator); ok {
		// Some providers may need/want to validate all params at once.
		err := v.ValidateAction(ctx, act)
		// If we get `ErrUnsupported`, we'll fall back to the field-by-field validation, this can happen if the provider implements the interface, but the backing implementation (e.g., external plugin) doesn't support it.
		if !errors.Is(err, ErrUnsupported) {
			return err
		}
	}

	knownParams := make([]string, 0, len(info.DynamicParams))
	for _, f := range info.DynamicParams {
		knownParams = append(knownParams, f.ParamID)
	}

	// Make sure we reject any params that are not expected.
	for paramID := range act.Params {
		if slices.Contains(knownParams, paramID) {
			continue
		}

		return &ActionParamError{
			ParamID: paramID,
			Err:     fmt.Errorf("unexpected param"),
		}
	}

	for _, param := range info.DynamicParams {
		val := "string(" + act.Params[param.ParamID] + ")"
		_, err = expr.Compile(val, expr.AsKind(reflect.String))
		if err != nil {
			return &ActionParamError{
				ParamID: param.ParamID,
				Err:     validation.WrapError(err),
			}
		}

		paramV, ok := p.(ParamValidator)
		if !ok {
			continue
		}

		err := paramV.ValidateParam(ctx, param.ParamID, act.Params[param.ParamID])
		if validation.IsClientError(err) {
			return &ActionParamError{
				ParamID: param.ParamID,
				Err:     err,
			}
		}
		if err != nil && !errors.Is(err, ErrUnsupported) {
			return err
		}
	}

	// Since we have no extra/unknown fields, and all required fields are valid, we've validated the Action.
	return nil
}
