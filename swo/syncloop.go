package swo

import (
	"context"
	"fmt"
	"time"

	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

// SyncLoop will loop the logical replication sequence until the number of changes reaches zero.
func (e *Execute) SyncLoop(ctx context.Context) {
	if e.err != nil {
		return
	}

	sync := func(ctx context.Context) (ok, pend int, err error) {
		srcTx, dstTx, err := e.syncTx(ctx, true)
		if err != nil {
			return 0, 0, fmt.Errorf("sync tx: %w", err)
		}
		defer srcTx.Rollback(ctx)
		defer dstTx.Rollback(ctx)

		ids, err := e.syncChanges(ctx, srcTx, dstTx)
		if err != nil {
			return 0, len(ids), fmt.Errorf("sync change log: %w", err)
		}

		err = srcTx.Commit(ctx)
		if err != nil {
			return len(ids), 0, fmt.Errorf("commit src: %w", err)
		}

		err = dstTx.Commit(ctx)
		if err != nil {
			return 0, len(ids), fmt.Errorf("commit dst: %w", err)
		}

		_, err = e.mainDBConn.Exec(ctx, "DELETE FROM change_log WHERE id = any($1)", sqlutil.IntArray(ids))
		if err != nil {
			return len(ids), 0, fmt.Errorf("update change log: %w", err)
		}

		return len(ids), 0, nil
	}

	for ctx.Err() == nil {
		// sync in a loop until DB is up-to-date
		s := time.Now()
		n, pend, err := sync(ctx)
		dur := time.Since(s)

		if pend > 0 {
			e.Progressf(ctx, "sync: %d rows pending", pend)
		}
		if err != nil {
			log.Log(ctx, err)
			e.Rollback()
			if n > 0 {
				e.err = fmt.Errorf("sync changes: sync failure (commit without record): %w", err)
				return
			}
			continue
		}
		e.Commit()

		if n > 10 {
			e.Progressf(ctx, "sync: %d rows replicated in %s", n, dur.Truncate(time.Millisecond))
			continue
		}

		return
	}

	e.err = fmt.Errorf("sync changes: %w", ctx.Err())
}
