package schedule

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

func validateTimeRange(prefix string, start, end time.Time) error {
	if !end.After(start) {
		return validation.NewFieldError(prefix+"End", fmt.Sprintf("must be after %sStart", prefix))
	}

	return nil
}
func validateWithinTimeRange(tPrefix, prefix string, tStart, tEnd, start, end time.Time) error {
	if tStart.Before(start) {
		return validation.NewFieldError(tPrefix+"Start", fmt.Sprintf("must not be before %sStart", prefix))
	}
	if !tStart.Before(end) {
		return validation.NewFieldError(tPrefix+"Start", fmt.Sprintf("must be before %sEnd", prefix))
	}
	if !tEnd.After(start) {
		return validation.NewFieldError(tPrefix+"End", fmt.Sprintf("must be after %sStart", prefix))
	}
	if tEnd.After(end) {
		return validation.NewFieldError(tPrefix+"End", fmt.Sprintf("must not be after %sEnd", prefix))
	}

	return nil
}

func (store *Store) validateShifts(ctx context.Context, fname string, max int, shifts []FixedShift, start, end time.Time) error {
	if len(shifts) > max {
		return validation.NewFieldError(fname, "too many shifts defined")
	}

	check, err := store.usr.UserExists(ctx)
	if err != nil {
		return err
	}

	for i, s := range shifts {
		prefix := fmt.Sprintf("%s[%d].", fname, i)

		err := validate.Many(
			validate.UUID(prefix+"UserID", s.UserID),
			validateTimeRange(prefix, s.Start, s.End),
			validateWithinTimeRange(prefix, "", s.Start, s.End, start, end),
		)
		if err != nil {
			return err
		}
		if !check.UserExistsString(s.UserID) {
			return validation.NewFieldError(prefix+"UserID", "user does not exist")
		}
	}

	return nil
}
