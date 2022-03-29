package metricsmanager

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

type State struct {
	V2 struct {
		LastLogTime time.Time
		LastLogID   int
	}
}

/*
	Theory of Operation:

	1. Aquire processing lock
	2. Get batch of oldest alert IDs (if cursor not blank, must be > cursor)
	3. Insert metrics for these alerts
	4. Set cursor to last inserted

*/

// UpdateAll will update the alert metrics table
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running metrics operations.")

	tx, lockState, err := db.lock.BeginTxWithState(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var alertIDs []int
	var lastLogTime time.Time
	var lastLogID int
	var state State
	err = lockState.Load(ctx, &state)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	var rows *sql.Rows
	rows, err = tx.StmtContext(ctx, db.scanLogs).QueryContext(ctx, state.V2.LastLogTime, state.V2.LastLogID)
	if err != nil {
		return fmt.Errorf("scan logs: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var alertID int
		err = rows.Scan(&alertID, &lastLogTime, &lastLogID)
		if err != nil {
			return fmt.Errorf("scan alert id: %w", err)
		}
		alertIDs = append(alertIDs, alertID)
	}

	if len(alertIDs) > 0 {
		_, err = tx.StmtContext(ctx, db.insertMetrics).ExecContext(ctx, sqlutil.IntArray(alertIDs))
		if err != nil {
			return fmt.Errorf("insert metrics: %w", err)
		}

		// update and save state
		state.V2.LastLogTime = lastLogTime
		state.V2.LastLogID = lastLogID
		err = lockState.Save(ctx, &state)
		if err != nil {
			return fmt.Errorf("save state: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
