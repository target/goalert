package metricsmanager

import (
	"context"
	"fmt"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

type State struct {
	V1 struct {
		NextAlertID int
	}
}

/*
	Theory of Operation:

	1. Aquire processing lock
	2. Look for recently closed alerts without a metrics entry
	3. If any, insert metrics for them and exit
	4. If no state, start scan from last closed alert id
	5. If state, resume scan until min closed alert id

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

	rows, err := tx.StmtContext(ctx, db.recentlyClosed).QueryContext(ctx)
	if err != nil {
		return fmt.Errorf("query recently closed alerts: %w", err)
	}
	defer rows.Close()

	var alertIDs []int
	for rows.Next() {
		var alertID int
		err = rows.Scan(&alertID)
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
		err = tx.Commit()
		if err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		return nil
	}

	var state State
	err = lockState.Load(ctx, &state)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}

	// fetch min alert id from db for later
	var minAlertID int
	err = tx.StmtContext(ctx, db.lowAlertID).QueryRowContext(ctx).Scan(&minAlertID)
	if err != nil {
		return fmt.Errorf("query min alert id: %w", err)
	}

	if state.V1.NextAlertID == 0 || state.V1.NextAlertID < minAlertID {
		// no state, or reset, set to the highest alert id from the db
		err = tx.StmtContext(ctx, db.highAlertID).QueryRowContext(ctx).Scan(&state.V1.NextAlertID)
		if err != nil {
			return fmt.Errorf("query high alert id: %w", err)
		}
	}

	// clamp min alert ID 500 below next
	if minAlertID < state.V1.NextAlertID-500 {
		minAlertID = state.V1.NextAlertID - 500
	}

	// fetch alerts to update
	rows, err = tx.StmtContext(ctx, db.scanAlerts).QueryContext(ctx, minAlertID, state.V1.NextAlertID)
	if err != nil {
		return fmt.Errorf("query alerts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var alertID int
		err = rows.Scan(&alertID)
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
	}

	// update and save state
	state.V1.NextAlertID = minAlertID - 1
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
