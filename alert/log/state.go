package alertlog

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

type LogState struct {
	LogID         int            `json:"log_ID"`
	LastStatus    sql.NullString `json:"last_status"`
	StatusDetails sql.NullString `json:"status_details"`
}

const maxBatch = 500

func (db *DB) FindOneLogState(ctx context.Context, logID int) (*LogState, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	row := db.findOneLogState.QueryRowContext(ctx, logID)
	var ls LogState
	err = row.Scan(&ls.LogID, &ls.LastStatus, &ls.StatusDetails)
	if err != nil {
		return nil, err
	}

	return &ls, nil
}

func (db *DB) FindManyLogStates(ctx context.Context, logIDs []int) ([]LogState, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.User)
	if err != nil {
		return nil, err
	}

	if len(logIDs) == 0 {
		return nil, nil
	}

	err = validate.Range("AlertLogIDs", len(logIDs), 1, maxBatch)
	if err != nil {
		return nil, err
	}

	rows, err := db.findManyLogStates.QueryContext(ctx, sqlutil.IntArray(logIDs))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []LogState
	var ls LogState
	for rows.Next() {
		err = rows.Scan(&ls.LogID, &ls.LastStatus, &ls.StatusDetails)
		if err != nil {
			return nil, err
		}
		result = append(result, ls)
	}

	return result, nil
}
