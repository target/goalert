package graphqlapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
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
func (a *App) ValidateDestination(ctx context.Context, fieldName string, dest *graphql2.DestinationInput) (err error) {
	cfg := config.FromContext(ctx)
	switch dest.Type {
	case destAlert:
		return nil
	case destTwilioSMS:
		phone := dest.FieldValue(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldPhoneNumber, err)
		}
		return nil
	case destTwilioVoice:
		phone := dest.FieldValue(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldPhoneNumber, err)
		}
		return nil
	case destSlackChan:
		chanID := dest.FieldValue(fieldSlackChanID)
		err := a.SlackStore.ValidateChannel(ctx, chanID)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldSlackChanID, err)
		}

		return nil
	case destSlackDM:
		userID := dest.FieldValue(fieldSlackUserID)
		if err := a.SlackStore.ValidateUser(ctx, userID); err != nil {
			return addDestFieldError(ctx, fieldName, fieldSlackUserID, err)
		}
		return nil
	case destSlackUG:
		ugID := dest.FieldValue(fieldSlackUGID)
		userErr := a.SlackStore.ValidateUserGroup(ctx, ugID)
		if userErr != nil {
			return addDestFieldError(ctx, fieldName, fieldSlackUGID, userErr)
		}

		chanID := dest.FieldValue(fieldSlackChanID)
		chanErr := a.SlackStore.ValidateChannel(ctx, chanID)
		if chanErr != nil {
			return addDestFieldError(ctx, fieldName, fieldSlackChanID, chanErr)
		}

		return nil
	case destSMTP:
		email := dest.FieldValue(fieldEmailAddress)
		err := validate.Email(fieldEmailAddress, email)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldEmailAddress, err)
		}
		return nil
	case destWebhook:
		url := dest.FieldValue(fieldWebhookURL)
		err := validate.AbsoluteURL(fieldWebhookURL, url)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldWebhookURL, err)
		}
		if !cfg.ValidWebhookURL(url) {
			return addDestFieldError(ctx, fieldName, fieldWebhookURL, validation.NewGenericError("url is not allowed by administator"))
		}
		return nil
	case destSchedule: // must be valid UUID and exist
		_, err := validate.ParseUUID(fieldScheduleID, dest.FieldValue(fieldScheduleID))
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldScheduleID, err)
		}

		_, err = a.ScheduleStore.FindOne(ctx, dest.FieldValue(fieldScheduleID))
		if errors.Is(err, sql.ErrNoRows) {
			return addDestFieldError(ctx, fieldName, fieldScheduleID, validation.NewGenericError("schedule does not exist"))
		}
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldScheduleID, err)
		}

		return nil
	case destRotation: // must be valid UUID and exist
		rotID := dest.FieldValue(fieldRotationID)
		_, err := validate.ParseUUID(fieldRotationID, rotID)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldRotationID, err)
		}
		_, err = a.RotationStore.FindRotation(ctx, rotID)
		if errors.Is(err, sql.ErrNoRows) {
			return addDestFieldError(ctx, fieldName, fieldRotationID, validation.NewGenericError("rotation does not exist"))
		}
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldRotationID, err)
		}

		return nil
	case destUser: // must be valid UUID and exist
		userID := dest.FieldValue(fieldUserID)
		uid, err := validate.ParseUUID(fieldUserID, userID)
		if err != nil {
			return addDestFieldError(ctx, fieldName, fieldUserID, err)
		}
		check, err := a.UserStore.UserExists(ctx)
		if err != nil {
			return fmt.Errorf("get user existance checker: %w", err)
		}
		if !check.UserExistsUUID(uid) {
			return addDestFieldError(ctx, fieldName, fieldUserID, validation.NewGenericError("user does not exist"))
		}
		return nil
	}

	// unsupported destination type
	graphql.AddError(ctx, &gqlerror.Error{
		Message: "unsupported destination type",
		Path:    appendPath(ctx, fieldName+".type"),
		Extensions: map[string]interface{}{
			"code": graphql2.ErrorCodeInvalidInputValue,
		},
	})

	return errAlreadySet
}
