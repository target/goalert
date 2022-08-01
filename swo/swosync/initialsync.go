package swosync

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swoinfo"
	"github.com/target/goalert/util/sqlutil"
)

func insertRowsQuery(table swoinfo.Table) string {
	return fmt.Sprintf(`
		insert into %s
		select * from
		json_populate_recordset(null::%s, $1)
	`, sqlutil.QuoteID(table.Name()), sqlutil.QuoteID(table.Name()))
}

// InitialSync will insert all rows from the source database into the destination database.
//
// While doing so it will update the rowID maps to track the rows that have been inserted.
func (l *LogicalReplicator) InitialSync(ctx context.Context) error {
	srcTx, err := l.srcConn.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.Serializable,
		DeferrableMode: pgx.Deferrable,
		AccessMode:     pgx.ReadOnly,
	})
	if err != nil {
		return fmt.Errorf("begin src tx: %w", err)
	}
	defer srcTx.Rollback(ctx)

	_, err = srcTx.Exec(ctx, txInProgressLock)
	if err != nil {
		return fmt.Errorf("lock tx: %w", err)
	}

	dstTx, err := l.dstConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin dst tx: %w", err)
	}
	defer dstTx.Rollback(ctx)

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

	insertSQL := insertRowsQuery(table)

	var insertRows []json.RawMessage
	var inserted int
	for rows.Next() {
		var id string
		var rowData json.RawMessage
		if err := rows.Scan(&id, &rowData); err != nil {
			return 0, fmt.Errorf("scan: %w", err)
		}
		insertRows = append(insertRows, rowData)
		l.dstRows.Set(changeID{table.Name(), id})

		if len(insertRows) < 10000 {
			continue
		}

		l.printf(ctx, "sync %s: %d/%d", table.Name(), inserted, count)
		_, err := dstTx.Exec(ctx, insertSQL, insertRows)
		if err != nil {
			return 0, fmt.Errorf("insert: %w", err)
		}
		inserted += len(insertRows)
		insertRows = insertRows[:0]
	}

	if len(insertRows) > 0 {
		_, err := dstTx.Exec(ctx, insertSQL, insertRows)
		if err != nil {
			return 0, fmt.Errorf("insert: %w", err)
		}
	}

	return count, nil
}
