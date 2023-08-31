package swosync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/util/sqlutil"
)

const maxBatchSize = 1024 * 1024 // 1MB

// FullInitialSync will insert all rows from the source database into the destination database.
//
// While doing so it will update the rowID maps to track the rows that have been inserted.
func (l *LogicalReplicator) FullInitialSync(ctx context.Context) error {
	srcTx, err := l.srcConn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		DeferrableMode: pgx.Deferrable,
		AccessMode:     pgx.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("begin src tx: %w", err)
	}
	defer sqlutil.RollbackContext(ctx, "swo: full initial sync: src tx", srcTx)

	_, err = srcTx.Exec(ctx, txInProgressLock)
	if err != nil {
		return fmt.Errorf("lock tx: %w", err)
	}

	dstTx, err := l.dstConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin dst tx: %w", err)
	}
	defer sqlutil.RollbackContext(ctx, "swo: full initial sync: dst tx", dstTx)

	_, err = dstTx.Exec(ctx, "set constraints all deferred")
	if err != nil {
		return fmt.Errorf("defer constraints: %w", err)
	}

	for _, table := range l.tables {
		_, err := l.initialSyncTable(ctx, srcTx, dstTx, table)
		if err != nil {
			return fmt.Errorf("initial sync table %s: %w", table.Name(), err)
		}
	}

	err = srcTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit src tx: %w", err)
	}

	err = dstTx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("commit dst tx: %w", err)
	}

	_, err = l.dstConn.Exec(ctx, "vacuum analyze")
	if err != nil {
		return fmt.Errorf("vacuum analyze: %w", err)
	}

	return nil
}

func (l *LogicalReplicator) initialSyncTable(ctx context.Context, srcTx, dstTx pgx.Tx, table swoinfo.Table) (int64, error) {
	l.printf(ctx, "sync %s", table.Name())
	var count int64
	err := srcTx.QueryRow(ctx, fmt.Sprintf("select count(*) from %s", sqlutil.QuoteID(table.Name()))).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	rows, err := srcTx.Query(ctx, fmt.Sprintf("select id::text, to_jsonb(tbl_row) from %s as tbl_row", sqlutil.QuoteID(table.Name())))
	if err != nil {
		return 0, fmt.Errorf("select: %w", err)
	}
	defer rows.Close()

	insertSQL := table.InsertJSONRowsQuery(false)

	doneCh := make(chan error)
	rowCh := make(chan json.RawMessage)
	go func() {
		var inserted int
		var dataSize int
		var batch []json.RawMessage
	sendLoop:
		for {
			var row json.RawMessage
			select {
			case row = <-rowCh:
				if row == nil {
					break sendLoop
				}
			case <-ctx.Done():
				return
			}
			batch = append(batch, row)
			dataSize += len(row)
			if dataSize < maxBatchSize {
				continue
			}

			l.printf(ctx, "sync %s: %d/%d", table.Name(), inserted, count)
			_, err := dstTx.Exec(ctx, insertSQL, batch)
			if err != nil {
				doneCh <- fmt.Errorf("insert: %w", err)
				return
			}

			inserted += len(batch)
			dataSize = 0
			batch = batch[:0]
		}

		if len(batch) > 0 {
			l.printf(ctx, "sync %s: %d/%d", table.Name(), inserted, count)
			_, err := dstTx.Exec(ctx, insertSQL, batch)
			if err != nil {
				doneCh <- fmt.Errorf("insert: %w", err)
				return
			}
		}

		doneCh <- nil
	}()

	for rows.Next() {
		var id string
		var rowData json.RawMessage
		if err := rows.Scan(&id, &rowData); err != nil {
			return 0, fmt.Errorf("scan: %w", err)
		}
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case rowCh <- rowData:
			continue
		case err := <-doneCh:
			return 0, err
		}
	}

	close(rowCh)
	if err := <-doneCh; err != nil {
		return 0, err
	}
	return count, nil
}
