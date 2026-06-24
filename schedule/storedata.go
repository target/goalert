package schedule

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util/jsonutil"
	"github.com/target/goalert/util/sqlutil"
)

func (store *Store) scheduleData(ctx context.Context, tx *sql.Tx, scheduleID uuid.UUID) (*Data, error) {
	db := gadb.New(store.db)
	if tx != nil {
		db = db.WithTx(tx)
	}
	rawData, err := db.SchedFindData(ctx, scheduleID)
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
		defer sqlutil.Rollback(ctx, "schedule: update data", tx)
	}

	db := gadb.New(store.db).WithTx(tx)

	var rawData json.RawMessage
	// Select for update, if it does not exist try inserting, if that fails due to a race, re-try select for update
	rawData, err = db.SchedFindDataForUpdate(ctx, scheduleID)
	if err == sql.ErrNoRows {
		err = db.SchedInsertData(ctx, scheduleID)
		if isDataPkeyConflict(err) {
			// insert happened after orig. select for update and our subsequent insert, re-try select for update
			rawData, err = db.SchedFindDataForUpdate(ctx, scheduleID)
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

	err = db.SchedUpdateData(ctx, gadb.SchedUpdateDataParams{
		ScheduleID: scheduleID,
		Data:       rawData,
	})
	if err != nil {
		return err
	}

	if !externalTx {
		return tx.Commit()
	}

	return nil
}
