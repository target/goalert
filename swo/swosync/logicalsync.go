package swosync

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// LogicalSync will sync the source database to the destination database as fast as possible.
func (l *LogicalReplicator) LogicalSync(ctx context.Context) error { return l.doSync(ctx, false) }

// FinalSync will sync the source database to the destination database, using the stop-the-world lock
// and updating switchover_state to use_next_db.
func (l *LogicalReplicator) FinalSync(ctx context.Context) error { return l.doSync(ctx, true) }

func (l *LogicalReplicator) doSync(ctx context.Context, final bool) error {
	b := new(pgx.Batch)
	if final {
		b.Queue(`begin`)
	} else {
		b.Queue(`begin isolation level serializable read only deferrable`)
	}
	b.Queue(txInProgressLock)
	if final {
		// stop-the-world lock before reads
		b.Queue(txStopTheWorld)
	}

	seqSync := NewSequenceSync(l.seqNames)
	seqSync.AddBatchReads(b)

	tblSync := NewTableSync(l.tables)
	tblSync.AddBatchChangeRead(b)

	res := l.srcConn.SendBatch(ctx, b)
	_, err := res.Exec() // begin tx
	if err != nil {
		return fmt.Errorf("read changes: begin tx: %w", err)
	}
	defer func(srcConn *pgx.Conn) { _, _ = srcConn.Exec(ctx, `rollback`) }(l.srcConn)

	// in-progress lock & check
	_, err = res.Exec()
	if err != nil {
		return fmt.Errorf("read changes: set tx timeout: %w", err)
	}

	if final {
		// stop-the-world lock before reads
		_, err = res.Exec()
		if err != nil {
			return fmt.Errorf("read changes: stop-the-world lock: %w", err)
		}
	}

	err = seqSync.ScanBatchReads(res)
	if err != nil {
		return fmt.Errorf("read changes: scan seqs: %w", err)
	}

	err = tblSync.ScanBatchChangeRead(res)
	if err != nil {
		return fmt.Errorf("read changes: scan changes: %w", err)
	}
	res.Close()

	var readRows pgx.Batch
	tblSync.AddBatchRowReads(&readRows)
	if readRows.Len() > 0 {
		res = l.srcConn.SendBatch(ctx, &readRows)
		err = tblSync.ScanBatchRowReads(res)
		if err != nil {
			return fmt.Errorf("read changes: scan rows: %w", err)
		}
		res.Close()
	}

	var applyChanges pgx.Batch
	applyChanges.Queue("begin")
	applyChanges.Queue("set constraints all deferred")
	seqSync.AddBatchWrites(&applyChanges)
	tblSync.AddBatchWrites(&applyChanges)
	applyChanges.Queue("commit")
	if final {
		// re-enable triggers in destination DB
		for _, t := range l.tables {
			applyChanges.Queue(fmt.Sprintf(`alter table %s enable trigger user`, sqlutil.QuoteID(t.Name())))
		}
	}
	err = l.dstConn.SendBatch(ctx, &applyChanges).Close()
	if err != nil {
		_, _ = l.dstConn.Exec(ctx, `rollback`)
		return fmt.Errorf("apply changes: %w", err)
	}

	var finish pgx.Batch
	if final {
		// world is stopped, changes in new DB, triggers enabled, so we can safely update switchover_state
		finish.Queue("update switchover_state set current_state = 'use_next_db' where current_state = 'in_progress'")
	}
	finish.Queue("commit")
	err = l.srcConn.SendBatch(ctx, &finish).Close()
	if err != nil {
		return fmt.Errorf("commit sync read: %w", err)
	}

	_, err = tblSync.ExecDeleteChanges(ctx, l.srcConn)
	if !final {
		return err
	}

	if err != nil {
		// log but don't return error in final since switchover is complete
		log.Log(ctx, err)
	}
	return nil
}
