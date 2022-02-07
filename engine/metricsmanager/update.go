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
	err := db.update(ctx)
	return err
}

func (db *DB) update(ctx context.Context) error {
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

	var state State
	var stateData []byte
	err = tx.StmtContext(ctx, db.findCurrentState).QueryRowContext(ctx).Scan(&stateData)
	if err != nil {
		return fmt.Errorf("get state: %w", err)
	}

	if len(stateData) > 0 {
		err = json.Unmarshal(stateData, &state)
		if err != nil {
			return fmt.Errorf("unmarshal state: %w", err)
		}
	}

	if state.MaxAlertID == 0 {
		err = tx.StmtContext(ctx, db.findMaxAlertID).QueryRowContext(ctx).Scan(&state.MaxAlertID)
		if err != nil {
			return fmt.Errorf("get max alertID: %w", err)
		}
	}
	
	return tx.Commit()
}
