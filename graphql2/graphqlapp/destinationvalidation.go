package graphqlapp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// errAlreadySet is returned when a field error is already set on the context.
var errAlreadySet = errors.New("error already set")

// errSkipHandler is a response interceptor that will skip errors with the "skip" extension set to true.
//
// This is used to ensure that errors are not returned to the client when they are not relevant (like errAlreadySet)
// and are only used for signaling to other parts of the system.
type errSkipHandler struct{}

var (
	_ graphql.ResponseInterceptor = errSkipHandler{}
	_ graphql.HandlerExtension    = errSkipHandler{}
)

func (errSkipHandler) ExtensionName() string { return "ErrorSkipHandler" }

func (errSkipHandler) Validate(schema graphql.ExecutableSchema) error { return nil }

func (errSkipHandler) InterceptResponse(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
	resp := next(ctx)

	filteredErrors := resp.Errors[:0]
	for _, err := range resp.Errors {
		if err.Extensions == nil {
			filteredErrors = append(filteredErrors, err)
			continue
		}

		skip, ok := err.Extensions["skip"].(bool)
		if !ok || !skip {
			filteredErrors = append(filteredErrors, err)
			continue
		}

		// skip this error
	}

	resp.Errors = filteredErrors

	return resp
}

// errReason will return the reason of the error, or the error message if it is not a field error.
func errReason(err error) string {
	var fErr validation.FieldError
	if errors.As(err, &fErr) {
		return fErr.Reason()
	}

	return err.Error()
}

// appendPath will append a field to the current path, taking into account array indices.
func appendPath(ctx context.Context, field string) ast.Path {
	p := graphql.GetPath(ctx)
	parentParts := strings.Split(field, ".")
	for _, part := range parentParts {
		if part == "" {
			continue
		}
		index, err := strconv.Atoi(part)
		if err == nil {
			p = append(p, ast.PathIndex(index))
			continue
		}

		p = append(p, ast.PathName(part))
	}

	return p
}

// addDestFieldError will add a destination field error to the current request, and return
// the original error if it is not a destination field validation error.
func addDestFieldError(ctx context.Context, parentField, fieldID string, err error) error {
	if err == nil {
		return nil
	}
	if permission.IsPermissionError(err) {
		// request level, return as is
		return err
	}
	if !validation.IsClientError(err) {
		// internal error, return as is
		return err
	}

	graphql.AddError(ctx, &gqlerror.Error{
		Message: errReason(err),
		Path:    appendPath(ctx, parentField),
		Extensions: map[string]interface{}{
			"code":    graphql2.ErrorCodeInvalidDestFieldValue,
			"fieldID": fieldID,
		},
	})

	return errAlreadySet
}

func addInputError(ctx context.Context, err error) {
	field := err.(validation.FieldError).Field()

	graphql.AddError(ctx, &gqlerror.Error{
		Message: errReason(err),
		Path:    appendPath(ctx, field),
		Extensions: map[string]interface{}{
			"code": graphql2.ErrorCodeInvalidInputValue,
		},
	})
}

// ValidateDestination will validate a destination input.
//
// In the future this will be a call to the plugin system.
func (a *App) ValidateDestination(ctx context.Context, fieldName string, dest *gadb.DestV1) (err error) {
	switch dest.Type {
	case destTwilioSMS:
		phone := dest.Arg(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldPhoneNumber, err)
		}
		return nil
	case destTwilioVoice:
		phone := dest.Arg(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldPhoneNumber, err)
		}
		return nil
	case destSMTP:
		email := dest.Arg(fieldEmailAddress)
		err := validate.Email(fieldEmailAddress, email)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldEmailAddress, err)
		}
		return nil
	}

	err = a.DestReg.ValidateDest(ctx, *dest)
	if errors.Is(err, nfydest.ErrUnknownType) {
		message := fmt.Sprintf("unsupported destination type: %s", dest.Type)
		if dest.Type == "" {
			message = "destination type is required"
		}

		// unsupported destination type
		graphql.AddError(ctx, &gqlerror.Error{
			Message: message,
			Path:    appendPath(ctx, fieldName+".type"),
			Extensions: map[string]interface{}{
				"code": graphql2.ErrorCodeInvalidInputValue,
			},
		})

		return errAlreadySet
	}

	var argErr *nfydest.DestArgError
	if errors.As(err, &argErr) {
		return addDestFieldError(ctx, fieldName+".args", argErr.FieldID, argErr.Err)
	}

	if err != nil {
		// internal error
		return err
	}

	return nil
}
