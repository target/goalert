package schedule

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

func validateShifts(fname string, max int, shifts []FixedShift) error {
	if len(shifts) > max {
		return validation.NewFieldError(fname, "too many shifts defined")
	}

	for i, s := range shifts {
		err := validate.UUID(fmt.Sprintf("%s[%d].UserID", fname, i), s.UserID)
		if err != nil {
			return err
		}
	}

	return nil
}

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

	var data ScheduleData
	if len(rawData) > 0 {
		err = json.Unmarshal(rawData, &data)
		if err != nil {
			return nil, err
		}
	}

	return data.V1.TemporarySchedules, nil
}
func (store *Store) updateFixedShifts(ctx context.Context, tx *sql.Tx, scheduleID string, apply func(data *ScheduleData) error) error {
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
	err = tx.StmtContext(ctx, store.findUpdData).QueryRowContext(ctx, scheduleID).Scan(&rawData)
	if err == sql.ErrNoRows {
		err = nil
	}
	if err != nil {
		return err
	}

	var data ScheduleData
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
func (store *Store) SetFixedShifts(ctx context.Context, tx *sql.Tx, scheduleID string, start, end time.Time, shifts []FixedShift) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("ScheduleID", scheduleID),
		validateShifts("Shifts", 500, shifts),
	)
	if err != nil {
		return err
	}

	return store.updateFixedShifts(ctx, tx, scheduleID, func(data *ScheduleData) error {
		data.V1.TemporarySchedules = setFixedShifts(data.V1.TemporarySchedules, start, end, shifts)
		return nil
	})
}
func (store *Store) ResetFixedShifts(ctx context.Context, tx *sql.Tx, scheduleID string, start, end time.Time) error {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return err
	}

	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return err
	}

	return store.updateFixedShifts(ctx, tx, scheduleID, func(data *ScheduleData) error {
		data.V1.TemporarySchedules = deleteFixedShifts(data.V1.TemporarySchedules, start, end)
		return nil
	})
}
