package swo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/target/goalert/swo/swogrp"
)

//go:embed changelog.sql
var changelogQuery string

func (e *Execute) exec(ctx context.Context, conn pgxQueryer, query string) {
	if e.err != nil {
		return
	}

	_, err := conn.Exec(ctx, query)
	if err != nil {
		e.err = fmt.Errorf("%s: %w", query, err)
		return
	}
}

func (e *Execute) readErr() error {
	err := e.err
	e.err = nil
	return err
}

// EnableChangeLog enables DB change tracking by creating a change_log table that
// records table and row IDs for each INSERT, UPDATE, or DELETE.
func (e *Execute) EnableChangeLog(ctx context.Context) {
	if e.err != nil {
		return
	}

	swogrp.Progressf(ctx, "enabling change log")
	e.exec(ctx, e.mainDBConn, changelogQuery)

	// create triggers for all tables
	for _, table := range e.tables {
		if table.SkipSync() {
			continue
		}
		query := fmt.Sprintf(`
			CREATE TRIGGER %s AFTER INSERT OR UPDATE OR DELETE ON %s
			FOR EACH ROW EXECUTE PROCEDURE fn_process_change_log()
		`, table.QuotedChangeTriggerName(), table.QuotedName())
		e.exec(ctx, e.mainDBConn, query)
	}

	e.exec(ctx, e.mainDBConn,
		"update switchover_state set current_state = 'in_progress' where current_state = 'idle'")

	if e.err != nil {
		e.err = fmt.Errorf("enable change log: %w", e.err)
	}
}
