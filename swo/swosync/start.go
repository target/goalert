package swosync

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/util/sqlutil"
)

//go:embed changelog.sql
var changelogQuery string

func triggerName(table string) string {
	return sqlutil.QuoteID(fmt.Sprintf("zz_99_change_log_%s", table))
}

// Start instruments and begins tracking changes to the DB.
func (l *LogicalReplicator) Start(ctx context.Context) error {
	l.printf(ctx, "enabling logical replication...")
	_, err := l.srcConn.Exec(ctx, changelogQuery)
	if err != nil {
		return fmt.Errorf("create change_log and fn: %w", err)
	}

	l.tables, err = swoinfo.ScanTables(ctx, l.srcConn)
	if err != nil {
		return fmt.Errorf("scan tables: %w", err)
	}

	l.seqNames, err = swoinfo.ScanSequences(ctx, l.srcConn)
	if err != nil {
		return fmt.Errorf("scan sequences: %w", err)
	}

	for _, table := range l.tables {
		// create change trigger in source DB
		chgTrigQuery := fmt.Sprintf(`
			CREATE TRIGGER %s AFTER INSERT OR UPDATE OR DELETE ON %s
			FOR EACH ROW EXECUTE PROCEDURE fn_process_change_log()
		`, triggerName(table.Name()), sqlutil.QuoteID(table.Name()))
		_, err := l.srcConn.Exec(ctx, chgTrigQuery)
		if err != nil {
			return fmt.Errorf("create change trigger for %s: %w", table.Name(), err)
		}

		// disable triggers in destination DB
		disableTrigQuery := fmt.Sprintf(`ALTER TABLE %s DISABLE TRIGGER USER`, sqlutil.QuoteID(table.Name()))
		_, err = l.dstConn.Exec(ctx, disableTrigQuery)
		if err != nil {
			return fmt.Errorf("disable trigger for %s: %w", table.Name(), err)
		}
	}

	err = swodb.New(l.srcConn).EnableChangeLogTriggers(ctx)
	if err != nil {
		return fmt.Errorf("enable change log triggers: %w", err)
	}

	// wait for in-flight transactions to finish
	l.printf(ctx, "waiting for in-flight transactions to finish")

	db := swodb.New(l.srcConn)

	now, err := db.Now(ctx)
	if err != nil {
		return fmt.Errorf("wait for active tx: get current time: %w", err)
	}

	for {
		n, err := db.ActiveTxCount(ctx, now)
		if err != nil {
			return fmt.Errorf("wait for active tx: get active tx count: %w", err)
		}
		if n == 0 {
			break
		}

		l.printf(ctx, "waiting for %d transaction(s) to finish", n)
		err = ctxSleep(ctx, time.Second)
		if err != nil {
			return fmt.Errorf("wait for active tx: sleep: %w", err)
		}
	}

	return nil
}

func ctxSleep(ctx context.Context, dur time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(dur):
	}
	return nil
}
