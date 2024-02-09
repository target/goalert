package graphqlapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

type FieldValueError struct {
	FieldID string `json:"fieldID"`
	Message string `json:"message"`
}

type DestinationValidationError struct {
	Type        string            `json:"type"` // always "DestinationValidationError"
	FieldErrors []FieldValueError `json:"fieldErrors"`
}

func (e *DestinationValidationError) Error() string {
	return "DestinationValidationError"
}
func (e *DestinationValidationError) ClientError() bool { return true }

func newDestErr(errs ...error) error {
	var destErr DestinationValidationError
	destErr.Type = "DestinationValidationError"
	for _, err := range errs {
		if f, ok := err.(validation.FieldError); ok {
			destErr.FieldErrors = append(destErr.FieldErrors, FieldValueError{
				FieldID: f.Field(),
				Message: f.Reason(),
			})
			continue
		}

		// non-field error, just return the bunch
		return errors.Join(errs...)
	}

	return &destErr
}

// ValidateDestination will validate a destination input.
//
// In the future this will be a call to the plugin system.
func (a *App) ValidateDestination(ctx context.Context, dest *graphql2.DestinationInput) error {
	cfg := config.FromContext(ctx)
	switch dest.Type {
	case destTwilioSMS:
		phone := dest.FieldValue(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return newDestErr(err)
		}
		return nil
	case destTwilioVoice:
		phone := dest.FieldValue(fieldPhoneNumber)
		err := validate.Phone(fieldPhoneNumber, phone)
		if err != nil {
			return newDestErr(err)
		}
		return nil
	case destSlackChan:
		chanID := dest.FieldValue(fieldSlackChanID)
		err := a.SlackStore.ValidateChannel(ctx, fieldSlackChanID, chanID)
		if err != nil {
			return newDestErr(err)
		}

		return nil
	case destSlackDM:
		userID := dest.FieldValue(fieldSlackUserID)
		err := a.SlackStore.ValidateUser(ctx, fieldSlackUserID, userID)
		if err != nil {
			return newDestErr(err)
		}
		return nil
	case destSlackUG:
		ugID := dest.FieldValue(fieldSlackUGID)
		chanID := dest.FieldValue(fieldSlackChanID)
		err := a.SlackStore.ValidateUserGroup(ctx, fieldSlackUGID, ugID)
		if err != nil {
			return newDestErr(err)
		}

		err = a.SlackStore.ValidateChannel(ctx, fieldSlackChanID, chanID)
		if err != nil {
			return newDestErr(err)
		}

		return nil
	case destSMTP:
		email := dest.FieldValue(fieldEmailAddress)
		err := validate.Email(fieldEmailAddress, email)
		if err != nil {
			return newDestErr(err)
		}
		return nil
	case destWebhook:
		url := dest.FieldValue(fieldWebhookURL)
		err := validate.AbsoluteURL(fieldWebhookURL, url)
		if err != nil {
			return newDestErr(err)
		}
		if !cfg.ValidWebhookURL(url) {
			return newDestErr(validation.NewFieldError(fieldWebhookURL, "url is not allowed by administator"))
		}
		return nil
	case destSchedule: // must be valid UUID and exist
		_, err := validate.ParseUUID(fieldScheduleID, dest.FieldValue(fieldScheduleID))
		if err != nil {
			return newDestErr(err)
		}

		_, err = a.ScheduleStore.FindOne(ctx, dest.FieldValue(fieldScheduleID))
		if errors.Is(err, sql.ErrNoRows) {
			return newDestErr(validation.NewFieldError(fieldScheduleID, "schedule does not exist"))
		}

		return err // return any other error
	case destRotation: // must be valid UUID and exist
		rotID := dest.FieldValue(fieldRotationID)
		_, err := validate.ParseUUID(fieldRotationID, rotID)
		if err != nil {
			return newDestErr(err)
		}
		_, err = a.RotationStore.FindRotation(ctx, rotID)
		if errors.Is(err, sql.ErrNoRows) {
			return newDestErr(validation.NewFieldError(fieldRotationID, "rotation does not exist"))
		}

		return err // return any other error

	case destUser: // must be valid UUID and exist
		userID := dest.FieldValue(fieldUserID)
		uid, err := validate.ParseUUID(fieldUserID, userID)
		if err != nil {
			return newDestErr(err)
		}
		check, err := a.UserStore.UserExists(ctx)
		if err != nil {
			return fmt.Errorf("get user existance checker: %w", err)
		}
		if !check.UserExistsUUID(uid) {
			return newDestErr(validation.NewFieldError(fieldUserID, "user does not exist"))
		}
		return nil
	}

	return fmt.Errorf("unsupported destination type: %s", dest.Type)
}
