package swo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

func (m *Manager) DoReset(ctx context.Context) error {
	err := m.withConnFromOld(ctx, ResetOldDB)
	if err != nil {
		return fmt.Errorf("reset old db: %w", err)
	}

	err = m.withConnFromNew(ctx, ResetNewDB)
	if err != nil {
		return fmt.Errorf("reset new db: %w", err)
	}

	return nil
}

// ResetNewDB will reset the new database to a clean state.
func ResetNewDB(ctx context.Context, conn *pgx.Conn) error {
	err := SwitchOverExecLock(ctx, conn)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer UnlockConn(ctx, conn)

	_, err = conn.Exec(ctx, "update switchover_state set current_state = 'idle' where current_state = 'in_progress'")
	if err != nil {
		return fmt.Errorf("set state to idle: %w", err)
	}

	tables, err := ScanTables(ctx, conn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	// truncate sync tables
	for _, table := range tables {
		if table.SkipSync() {
			continue
		}

		_, err = conn.Exec(ctx, fmt.Sprintf("truncate %s", table.QuotedName()))
		if err != nil {
			return fmt.Errorf("truncate %s: %w", table.QuotedName(), err)
		}
	}

	// drop the change_log table
	_, err = conn.Exec(ctx, "drop table if exists change_log")
	if err != nil {
		return fmt.Errorf("drop change_log: %w", err)
	}

	return nil
}

// ResetOldDB will reset the old database to a clean state.
//
// It will remove all change triggers and cleanup switchover data.
func ResetOldDB(ctx context.Context, conn *pgx.Conn) error {
	err := SwitchOverExecLock(ctx, conn)
	if err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer UnlockConn(ctx, conn)

	_, err = conn.Exec(ctx, "update switchover_state set current_state = 'idle' where current_state = 'in_progress'")
	if err != nil {
		return fmt.Errorf("set state to idle: %w", err)
	}

	tables, err := ScanTables(ctx, conn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	// drop change triggers
	for _, table := range tables {
		if table.SkipSync() {
			continue
		}

		_, err = conn.Exec(ctx, fmt.Sprintf("drop trigger if exists %s on %s", table.QuotedChangeTriggerName(), table.QuotedName()))
		if err != nil {
			return fmt.Errorf("drop trigger %s: %w", table.QuotedChangeTriggerName(), err)
		}

		_, err = conn.Exec(ctx, fmt.Sprintf("drop trigger if exists %s on %s", table.QuotedLockTriggerName(), table.QuotedName()))
		if err != nil {
			return fmt.Errorf("drop trigger %s: %w", table.QuotedChangeTriggerName(), err)
		}

	}

	// TODO: ensure no deps get missed
	_, err = conn.Exec(ctx, "DROP FUNCTION IF EXISTS fn_switchover_change_log_lock()")
	if err != nil {
		return fmt.Errorf("drop fn_switchover_change_log_lock: %w", err)
	}

	_, err = conn.Exec(ctx, "DROP FUNCTION IF EXISTS fn_process_change_log()")
	if err != nil {
		return fmt.Errorf("drop fn_process_change_log: %w", err)
	}

	// drop the change_log table
	_, err = conn.Exec(ctx, "drop table if exists change_log")
	if err != nil {
		return fmt.Errorf("drop change_log: %w", err)
	}

	return nil
}
