package schedule

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// FixedShiftsPerGroupLimit is the maximum number of shifts that can be configured for a single group at a time.
const FixedShiftsPerGroupLimit = 500

func (store *Store) validateShifts(ctx context.Context, fname string, max int, shifts []FixedShift) error {
	if len(shifts) > max {
		return validation.NewFieldError(fname, "too many shifts defined")
	}

	check, err := store.usr.UserExists(ctx)
	if err != nil {
		return err
	}
	defer check.Done()

	for i, s := range shifts {
		err := validate.UUID(fmt.Sprintf("%s[%d].UserID", fname, i), s.UserID)
		if err != nil {
			return err
		}
		if !check.UserExistsString(s.UserID) {
			return validation.NewFieldError(fmt.Sprintf("%s[%d].UserID", fname, i), "user does not exist")
		}
	}

	return nil
}

// FixedShiftGroups will return the current set for the provided scheduleID.
func (store *Store) FixedShiftGroups(ctx context.Context, tx *sql.Tx, scheduleID string) ([]FixedShiftGroup, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}

	stmt := store.findData
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	var rawData json.RawMessage
	err = stmt.QueryRowContext(ctx, scheduleID).Scan(&rawData)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	var data Data
	if len(rawData) > 0 {
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return nil, err
		}
	}

	check, err := store.usr.UserExists(ctx)
	if err != nil {
		return nil, err
	}
	defer check.Done()

	// omit shifts for non-existant users
	for i, tmp := range data.V1.TemporarySchedules {
		shifts := tmp.Shifts[:0]
		for _, shift := range tmp.Shifts {
			if !check.UserExistsString(shift.UserID) {
				continue
			}
			shifts = append(shifts, shift)
		}
		tmp.Shifts = shifts
		data.V1.TemporarySchedules[i] = tmp
	}

	return data.V1.TemporarySchedules, nil
}

func isDataPkeyConflict(err error) bool {
	dbErr := sqlutil.MapError(err)
	if dbErr == nil {
		return false
	}
	return dbErr.ConstraintName == "schedule_data_pkey"
}
func (store *Store) updateFixedShifts(ctx context.Context, tx *sql.Tx, scheduleID string, apply func(data *Data) error) error {
	var err error
	externalTx := tx != nil
	if !externalTx {
		tx, err = store.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
	}

	var rawData json.RawMessage
	// Select for update, if it does not exist try inserting, if that fails due to a race, re-try select for update
	err = tx.StmtContext(ctx, store.findUpdData).QueryRowContext(ctx, scheduleID).Scan(&rawData)
	if err == sql.ErrNoRows {
		_, err = tx.StmtContext(ctx, store.insertData).ExecContext(ctx, scheduleID)
		if isDataPkeyConflict(err) {
			// insert happened after orig. select for update and our subsequent insert, re-try select for update
			err = tx.StmtContext(ctx, store.findUpdData).QueryRowContext(ctx, scheduleID).Scan(&rawData)
		}
	}
	if err != nil {
		return err
	}

	var data Data
	if len(rawData) > 0 {
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return err
		}
	}

	err = apply(&data)
	if err != nil {
		return err
	}

	rawData, err = json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = tx.StmtContext(ctx, store.updateData).ExecContext(ctx, scheduleID, rawData)
	if err != nil {
		return err
	}

	if !externalTx {
		return tx.Commit()
	}

	return nil
}

// SetFixedShifts will cause the schedule to use only, and exactly the provided set of shifts between the provided start and end times.
func (store *Store) SetFixedShifts(ctx context.Context, tx *sql.Tx, scheduleID string, start, end time.Time, shifts []FixedShift) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("ScheduleID", scheduleID),
		store.validateShifts(ctx, "Shifts", FixedShiftsPerGroupLimit, shifts),
	)
	if err != nil {
		return err
	}

	return store.updateFixedShifts(ctx, tx, scheduleID, func(data *Data) error {
		data.V1.TemporarySchedules = setFixedShifts(data.V1.TemporarySchedules, start, end, shifts)
		return nil
	})
}

// ResetFixedShifts will clear out (or split, if needed) any defined fixed-shift groups that exist between the start and end time.
func (store *Store) ResetFixedShifts(ctx context.Context, tx *sql.Tx, scheduleID string, start, end time.Time) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return err
	}

	return store.updateFixedShifts(ctx, tx, scheduleID, func(data *Data) error {
		data.V1.TemporarySchedules = deleteFixedShifts(data.V1.TemporarySchedules, start, end)
		return nil
	})
}
