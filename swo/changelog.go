package swo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v4"
)

var (
	//go:embed changelog_table.sql
	changelogTable string

	//go:embed changelog_trigger.sql
	changelogTrigger string
)

func EnableChangeLog(ctx context.Context, conn *pgx.Conn) error {
	tables, err := ScanTables(ctx, conn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	_, err = conn.Exec(ctx, changelogTable)
	if err != nil {
		return fmt.Errorf("create change_log table: %w", err)
	}
	_, err = conn.Exec(ctx, changelogTrigger)
	if err != nil {
		return fmt.Errorf("create change_log AFTER trigger: %w", err)
	}
	_, err = conn.Exec(ctx, `insert into change_log(id,table_name,op,row_id) values(0,'','INIT',0)`)
	if err != nil {
		return fmt.Errorf("create change_log INIT row: %w", err)
	}

	// create triggers
	for _, table := range tables {
		if table.SkipSync() {
			continue
		}

		_, err = conn.Exec(ctx, fmt.Sprintf(`
			CREATE TRIGGER %s AFTER INSERT OR UPDATE OR DELETE ON %s
			FOR EACH ROW EXECUTE PROCEDURE fn_process_change_log()
		`, table.QuotedChangeTriggerName(), table.QuotedName()))
		if err != nil {
			return fmt.Errorf("create trigger %s: %w", table.QuotedChangeTriggerName(), err)
		}
	}

	_, err = conn.Exec(ctx, "update switchover_state set current_state = 'in_progress' where current_state = 'idle'")
	if err != nil {
		return fmt.Errorf("update switchover_state to in_progress: %w", err)
	}

	return nil
}
