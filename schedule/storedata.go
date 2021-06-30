package schedule

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/target/goalert/util/jsonutil"
)

func (store *Store) scheduleData(ctx context.Context, tx *sql.Tx, scheduleID uuid.UUID) (*Data, error) {
	stmt := store.findData
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	var rawData json.RawMessage
	err := stmt.QueryRowContext(ctx, scheduleID).Scan(&rawData)
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

	return &data, nil
}

func (store *Store) updateScheduleData(ctx context.Context, tx *sql.Tx, scheduleID uuid.UUID, apply func(data *Data) error) error {
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

	// preserve unknown fields
	rawData, err = jsonutil.Apply(rawData, data)
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
