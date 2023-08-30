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
		// LastLogTime is a cursor for processed alert_logs
		LastLogTime time.Time

		// LastLogID breaks ties for the LastLogTime cursor
		LastLogID int

		// LastMetricsDate is a cursor for processed alert_metrics
		LastMetricsDate time.Time
	}
}

func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}

	err = db.UpdateAlertMetrics(ctx)
	if err != nil {
		return err
	}

	return nil
}

// UpdateAlertMetrics will update the alert metrics table
/*
	Theory of Operation:

	1. Acquire processing lock
	2. Get batch of oldest alert IDs after cursor (use LastLogID as a tie breaker)
	3. Insert metrics for these alerts
	4. Set cursor to last inserted
	5. If none inserted, set cursor to upper time bound

*/
func (db *DB) UpdateAlertMetrics(ctx context.Context) error {
	log.Debugf(ctx, "Running alert_metrics operations.")

	tx, lockState, err := db.lock.BeginTxWithState(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer sqlutil.Rollback(ctx, "metrics manager", tx)

	var alertIDs []int
	var lastLogTime, boundNow time.Time
	var lastLogID int
	var state State
	err = lockState.Load(ctx, &state)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	err = tx.StmtContext(ctx, db.boundNow).QueryRowContext(ctx).Scan(&boundNow)
	if err != nil {
		return fmt.Errorf("select bound now: %w", err)
	}

	var rows *sql.Rows
	rows, err = tx.StmtContext(ctx, db.scanLogs).QueryContext(ctx, state.V2.LastLogTime, state.V2.LastLogID, boundNow)
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

	if len(alertIDs) == 0 {
		return nil
	}

	_, err = tx.StmtContext(ctx, db.insertMetrics).ExecContext(ctx, sqlutil.IntArray(alertIDs))
	if err != nil {
		return fmt.Errorf("insert metrics: %w", err)
	}

	// update state
	state.V2.LastLogTime = lastLogTime
	state.V2.LastLogID = lastLogID

	// save state
	err = lockState.Save(ctx, &state)
	if err != nil {
		return fmt.Errorf("save state: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
