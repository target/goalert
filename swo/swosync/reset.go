package swosync

import (
	"context"
	"fmt"
	"strings"

	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/util/sqlutil"
)

func (l *LogicalReplicator) Reset(ctx context.Context) error {
	l.printf(ctx, "disabling logical replication...")

	_, err := l.srcConn.Exec(ctx, ConnLockQuery)
	if err != nil {
		return fmt.Errorf("error locking source database: %w", err)
	}

	err = swodb.New(l.srcConn).DisableChangeLogTriggers(ctx)
	if err != nil {
		return fmt.Errorf("disable change log triggers: %w", err)
	}

	l.tables, err = swoinfo.ScanTables(ctx, l.srcConn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	var tableNames []string
	for _, table := range l.tables {
		// delete change trigger in source DB
		chgTrigQuery := fmt.Sprintf(`drop trigger if exists %s on %s`, triggerName(table.Name()), sqlutil.QuoteID(table.Name()))
		_, err := l.srcConn.Exec(ctx, chgTrigQuery)
		if err != nil {
			return fmt.Errorf("delete change trigger for %s: %w", table.Name(), err)
		}

		tableNames = append(tableNames, sqlutil.QuoteID(table.Name()))
	}

	// drop change_log table and func
	_, err = l.srcConn.Exec(ctx, `drop function if exists fn_process_change_log()`)
	if err != nil {
		return fmt.Errorf("drop fn_process_change_log: %w", err)
	}
	_, err = l.srcConn.Exec(ctx, `drop table if exists change_log`)
	if err != nil {
		return fmt.Errorf("drop change_log: %w", err)
	}

	l.printf(ctx, "clearing dest DB")
	_, err = l.dstConn.Exec(ctx, "truncate "+strings.Join(tableNames, ","))
	if err != nil {
		return fmt.Errorf("truncate tables: %w", err)
	}

	l.tables = nil
	l.seqNames = nil

	return nil
}
