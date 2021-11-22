package schedule

import (
	"fmt"
	"time"

	"github.com/target/goalert/user"
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

func (temp TemporarySchedule) validateShifts(checkUser user.ExistanceChecker) error {
	if len(temp.Shifts) > FixedShiftsPerTemporaryScheduleLimit {
		return validation.NewFieldError("Shifts", "too many shifts defined")
	}

	for i, s := range temp.Shifts {
		prefix := fmt.Sprintf("Shifts[%d].", i)

		err := validate.Many(
			validate.UUID(prefix+"UserID", s.UserID),
			validateTimeRange(prefix, s.Start, s.End),
			validateWithinTimeRange(prefix, "", s.Start, s.End, temp.Start, temp.End),
		)
		if err != nil {
			return err
		}
		if checkUser != nil && !checkUser.UserExistsString(s.UserID) {
			return validation.NewFieldError(prefix+"UserID", "user does not exist")
		}
	}

	return nil
}
