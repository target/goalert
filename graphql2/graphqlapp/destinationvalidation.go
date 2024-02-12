package graphqlapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

const (
	ErrCodeInvalidDestType  = "INVALID_DESTINATION_TYPE"
	ErrCodeInvalidDestValue = "INVALID_DESTINATION_FIELD_VALUE"
)

// addDestFieldError will add a destination field error to the current request, and return
// the original error if it is not a destination field validation error.
func addDestFieldError(ctx context.Context, parentField, fieldID string, err error) error {
	if permission.IsPermissionError(err) {
		// request level, return as is
		return err
	}
	if !validation.IsClientError(err) {
		// internal error, return as is
		return err
	}

	p := graphql.GetPath(ctx)
	p = append(p,
		ast.PathName(parentField),
		ast.PathName("values"), // DestinationInput.Values
		ast.PathName(fieldID),
	)

	graphql.AddError(ctx, &gqlerror.Error{
		Message: err.Error(),
		Path:    p,
		Extensions: map[string]interface{}{
			"code": ErrCodeInvalidDestValue,
		},
	})

	return nil
}

// ValidateDestination will validate a destination input.
//
// In the future this will be a call to the plugin system.
func (a *App) ValidateDestination(ctx context.Context, fieldName string, dest *graphql2.DestinationInput) (ok bool, err error) {
	cfg := config.FromContext(ctx)
	switch dest.Type {
	case destTwilioSMS:
		phone := dest.FieldValue(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldPhoneNumber, err)
		}
		return true, nil
	case destTwilioVoice:
		phone := dest.FieldValue(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldPhoneNumber, err)
		}
		return true, nil
	case destSlackChan:
		chanID := dest.FieldValue(fieldSlackChanID)
		err := a.SlackStore.ValidateChannel(ctx, chanID)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldSlackChanID, err)
		}

		return true, nil
	case destSlackDM:
		userID := dest.FieldValue(fieldSlackUserID)
		if err := a.SlackStore.ValidateUser(ctx, userID); err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldSlackUserID, err)
		}
		return true, nil
	case destSlackUG:
		ugID := dest.FieldValue(fieldSlackUGID)
		userErr := a.SlackStore.ValidateUserGroup(ctx, ugID)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldSlackUGID, userErr)
		}

		chanID := dest.FieldValue(fieldSlackChanID)
		chanErr := a.SlackStore.ValidateChannel(ctx, chanID)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldSlackChanID, chanErr)
		}

		return true, nil
	case destSMTP:
		email := dest.FieldValue(fieldEmailAddress)
		err := validate.Email(fieldEmailAddress, email)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldEmailAddress, err)
		}
		return true, nil
	case destWebhook:
		url := dest.FieldValue(fieldWebhookURL)
		err := validate.AbsoluteURL(fieldWebhookURL, url)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldWebhookURL, err)
		}
		if !cfg.ValidWebhookURL(url) {
			return false, addDestFieldError(ctx, fieldName, fieldWebhookURL, validation.NewGenericError("url is not allowed by administator"))
		}
		return true, nil
	case destSchedule: // must be valid UUID and exist
		_, err := validate.ParseUUID(fieldScheduleID, dest.FieldValue(fieldScheduleID))
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldScheduleID, err)
		}

		_, err = a.ScheduleStore.FindOne(ctx, dest.FieldValue(fieldScheduleID))
		if errors.Is(err, sql.ErrNoRows) {
			return false, addDestFieldError(ctx, fieldName, fieldScheduleID, validation.NewGenericError("schedule does not exist"))
		}
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldScheduleID, err)
		}

		return true, nil
	case destRotation: // must be valid UUID and exist
		rotID := dest.FieldValue(fieldRotationID)
		_, err := validate.ParseUUID(fieldRotationID, rotID)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldRotationID, err)
		}
		_, err = a.RotationStore.FindRotation(ctx, rotID)
		if errors.Is(err, sql.ErrNoRows) {
			return false, addDestFieldError(ctx, fieldName, fieldRotationID, validation.NewGenericError("rotation does not exist"))
		}
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldRotationID, err)
		}

		return true, nil
	case destUser: // must be valid UUID and exist
		userID := dest.FieldValue(fieldUserID)
		uid, err := validate.ParseUUID(fieldUserID, userID)
		if err != nil {
			return false, addDestFieldError(ctx, fieldName, fieldUserID, err)
		}
		check, err := a.UserStore.UserExists(ctx)
		if err != nil {
			return false, fmt.Errorf("get user existance checker: %w", err)
		}
		if !check.UserExistsUUID(uid) {
			return false, addDestFieldError(ctx, fieldName, fieldUserID, validation.NewGenericError("user does not exist"))
		}
		return true, nil
	}

	// unsupported destination type
	p := graphql.GetPath(ctx)
	p = append(p,
		ast.PathName(fieldName),
		ast.PathName("type"),
	)

	graphql.AddError(ctx, &gqlerror.Error{
		Message: "unsupported destination type",
		Path:    p,
		Extensions: map[string]interface{}{
			"code": ErrCodeInvalidDestType,
		},
	})

	return false, nil
}
