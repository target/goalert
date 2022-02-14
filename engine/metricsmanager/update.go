package metricsmanager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util/log"
)

type State struct {
	MaxAlertID int
}

// UpdateAll will update the alert metrics table
func (db *DB) UpdateAll(ctx context.Context) error {
	err := permission.LimitCheckAny(ctx, permission.System)
	if err != nil {
		return err
	}
	log.Debugf(ctx, "Running metrics operations.")

	tx, err := db.lock.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.StmtContext(ctx, db.setTimeout).ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("set timeout: %w", err)
	}

	// else

	// var minAlertID int
	// err = tx.StmtContext(ctx, db.findMinClosedAlertID).QueryRowContext(ctx).Scan(&minAlertID)
	// if err != nil {
	// 	return fmt.Errorf("get min closed alert id: %w", err)
	// }

	var state State
	var stateData []byte
	err = tx.StmtContext(ctx, db.findState).QueryRowContext(ctx).Scan(&stateData)
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}

	if len(stateData) > 0 {
		err = json.Unmarshal(stateData, &state)
		if err != nil {
			return fmt.Errorf("unmarshal state: %w", err)
		}
	}

	updateState := func() error {
		if state.MaxAlertID <= 0 {
			err = tx.StmtContext(ctx, db.findMaxAlertID).QueryRowContext(ctx).Scan(&state.MaxAlertID)
			if err != nil {
				return fmt.Errorf("get max alertID: %w", err)
			}
		}

		b, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("marshal state struct: %w", err)
		}

		_, err = tx.StmtContext(ctx, db.updateState).ExecContext(ctx, string(b))
		if err != nil {
			return fmt.Errorf("update state: %w", err)
		}
		return nil
	}

	err = updateState()
	if err != nil {
		return err
	}

	var recentAlertID, lowerBound, upperBound int
	err = tx.StmtContext(ctx, db.findRecentAlert).QueryRowContext(ctx).Scan(&recentAlertID)
	if err != nil {
		return fmt.Errorf("get recentl alert id: %w", err)
	}

	isUsingState := false
	if recentAlertID != 0 {
		lowerBound = recentAlertID - 3000
		upperBound = recentAlertID
	} else {
		lowerBound = state.MaxAlertID - 3000
		upperBound = state.MaxAlertID
		state.MaxAlertID = lowerBound
		isUsingState = true
	}

	_, err = tx.StmtContext(ctx, db.insertAlertMetrics).ExecContext(ctx, lowerBound, upperBound)
	if err != nil {
		return fmt.Errorf("insert alert metrics: %w", err)
	}

	if isUsingState {
		err := updateState()
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
