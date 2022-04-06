package swo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swogrp"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

func WaitForRunningTx(ctx context.Context, oldConn *pgx.Conn) error {
	var now time.Time
	err := oldConn.QueryRow(ctx, "select now()").Scan(&now)
	if err != nil {
		return fmt.Errorf("get current timestamp: %w", err)
	}

	for {
		var n int
		err = oldConn.QueryRow(ctx, "select count(*) from pg_stat_activity where state <> 'idle' and xact_start <= $1", now).Scan(&n)
		if err != nil {
			return fmt.Errorf("get running tx count: %w", err)
		}
		if n == 0 {
			break
		}

		swogrp.Progressf(ctx, "waiting for %d transaction(s) to finish", n)
		time.Sleep(time.Second)
	}

	return nil
}

func (m *Manager) DoExecute(ctx context.Context) error {
	/*
		- initial sync
		- loop until few changes
		- send proposal
		- loop until proposal is valid
		- execute proposal

	*/

	return m.withConnFromBoth(ctx, func(ctx context.Context, oldConn, newConn *pgx.Conn) error {
		swogrp.Progressf(ctx, "scanning tables...")
		tables, err := ScanTables(ctx, oldConn)
		if err != nil {
			return fmt.Errorf("scan tables: %w", err)
		}

		swogrp.Progressf(ctx, "enabling change log")
		err = EnableChangeLog(ctx, tables, oldConn)
		if err != nil {
			return fmt.Errorf("enable change log: %w", err)
		}

		swogrp.Progressf(ctx, "disabling triggers")
		err = DisableTriggers(ctx, tables, newConn)
		if err != nil {
			return fmt.Errorf("disable triggers: %w", err)
		}

		swogrp.Progressf(ctx, "waiting for in-flight transactions to finish")
		err = WaitForRunningTx(ctx, oldConn)
		if err != nil {
			return fmt.Errorf("wait for running tx: %w", err)
		}

		swogrp.Progressf(ctx, "performing initial sync")
		err = m.InitialSync(ctx, tables, oldConn, newConn)
		if err != nil {
			return fmt.Errorf("initial sync: %w", err)
		}

		swogrp.Progressf(ctx, "recording new DB state")
		rt, err := newRowTracker(ctx, tables, newConn)
		if err != nil {
			return fmt.Errorf("read row IDs: %w", err)
		}

		var lastNone bool
		for ctx.Err() == nil {
			// sync in a loop until DB is up-to-date
			s := time.Now()
			n, pend, err := LoopSync(ctx, rt, oldConn, newConn)
			dur := time.Since(s)

			if pend > 0 {
				lastNone = false
				swogrp.Progressf(ctx, "sync: %d rows pending", pend)
			}
			if err != nil {
				log.Log(ctx, err)
				rt.Rollback()
				if n > 0 {
					return fmt.Errorf("sync failure (commit without record): %w", err)
				}
				continue
			}
			rt.Commit()
			if n == 0 {
				if !lastNone {
					lastNone = true
					swogrp.Progressf(ctx, "sync: waiting for changes")
				}
				time.Sleep(100 * time.Millisecond)
			} else {
				lastNone = false
				swogrp.Progressf(ctx, "sync: %d rows replicated in %s", n, dur.Truncate(time.Millisecond))
			}
		}

		return errors.New("not implemented")
	})
}

// DisableTriggers will disable all triggers in the new DB.
func DisableTriggers(ctx context.Context, tables []Table, conn *pgx.Conn) error {
	for _, table := range tables {
		_, err := conn.Exec(ctx, fmt.Sprintf("ALTER TABLE %s DISABLE TRIGGER USER", table.QuotedName()))
		if err != nil {
			return fmt.Errorf("%s: %w", table.Name, err)
		}
	}

	return nil
}

func LoopSync(ctx context.Context, rt *rowTracker, srcConn, dstConn *pgx.Conn) (ok, pend int, err error) {
	srcTx, dstTx, err := syncTx(ctx, srcConn, dstConn)
	if err != nil {
		return 0, 0, fmt.Errorf("sync tx: %w", err)
	}
	defer srcTx.Rollback(ctx)
	defer dstTx.Rollback(ctx)

	ids, err := syncChangeLog(ctx, rt, srcTx, dstTx)
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

	_, err = srcConn.Exec(ctx, "DELETE FROM change_log WHERE id = any($1)", sqlutil.IntArray(ids))
	if err != nil {
		return len(ids), 0, fmt.Errorf("update change log: %w", err)
	}

	return len(ids), 0, nil
}

func FinalSync(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	return nil
}

func syncTx(ctx context.Context, srcConn, dstConn *pgx.Conn) (src, dst pgx.Tx, err error) {
	srcTx, err := srcConn.BeginTx(ctx, pgx.TxOptions{
		AccessMode:     pgx.ReadOnly,
		IsoLevel:       pgx.Serializable,
		DeferrableMode: pgx.Deferrable,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("begin src: %w", err)
	}

	dstTx, err := dstConn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		srcTx.Rollback(ctx)
		return nil, nil, fmt.Errorf("begin dst: %w", err)
	}

	return srcTx, dstTx, nil
}

func syncChangeLog(ctx context.Context, rt *rowTracker, srcTx, dstTx pgx.Tx) ([]int, error) {
	type rowID struct {
		table string
		id    string
	}

	var r rowID
	var changeIDs []int
	var changeID int
	changes := make(map[rowID]struct{})
	rowIDs := make(map[string][]string)
	_, err := srcTx.QueryFunc(ctx, "select id, table_name, row_id from change_log", nil, []interface{}{&changeID, &r.table, &r.id}, func(pgx.QueryFuncRow) error {
		if _, ok := changes[r]; ok {
			return nil
		}
		changes[r] = struct{}{}
		rowIDs[r.table] = append(rowIDs[r.table], r.id)
		changeIDs = append(changeIDs, changeID)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("fetch changes: %w", err)
	}
	if len(changes) == 0 {
		return nil, nil
	}

	// defer all constraints
	_, err = dstTx.Exec(ctx, "SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		return changeIDs, fmt.Errorf("defer constraints: %w", err)
	}

	type pendingDelete struct {
		query string
		idArg interface{}
		count int
	}
	var deletes []pendingDelete

	// go in insert order for fetching updates/inserts, note deleted rows
	for _, table := range rt.tables {
		if table.SkipSync() {
			continue
		}

		if len(rowIDs[table.Name]) == 0 {
			continue
		}

		sd, err := rt.fetch(ctx, table, srcTx, rowIDs[table.Name])
		if err != nil {
			return changeIDs, fmt.Errorf("fetch changed rows: %w", err)
		}
		if len(sd.toDelete) > 0 {
			deletes = append(deletes, pendingDelete{table.DeleteRowsQuery(), table.IDs(sd.toDelete), len(sd.toDelete)})
		}

		err = rt.apply(ctx, dstTx, table.UpdateRowsQuery(), sd.toUpdate)
		if err != nil {
			return changeIDs, fmt.Errorf("apply updates: %w", err)
		}

		err = rt.apply(ctx, dstTx, table.InsertRowsQuery(), sd.toInsert)
		if err != nil {
			return changeIDs, fmt.Errorf("apply inserts: %w", err)
		}
	}

	// handle pendingDeletes in reverse table order
	for i := len(deletes) - 1; i >= 0; i-- {
		t, err := dstTx.Exec(ctx, deletes[i].query, deletes[i].idArg)
		if err != nil {
			return changeIDs, fmt.Errorf("delete rows: %w", err)
		}
		if t.RowsAffected() != int64(deletes[i].count) {
			return changeIDs, fmt.Errorf("delete rows: got %d != expected %d", t.RowsAffected(), deletes[i].count)
		}
	}

	return changeIDs, nil
}

func (rt *rowTracker) apply(ctx context.Context, dstTx pgx.Tx, q string, rows []syncRow) error {
	if len(rows) == 0 {
		return nil
	}

	var rowsData []json.RawMessage
	for _, row := range rows {
		rowsData = append(rowsData, row.data)
	}

	data, err := json.Marshal(rowsData)
	if err != nil {
		return fmt.Errorf("marshal rows: %w", err)
	}
	t, err := dstTx.Exec(ctx, q, data)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}
	if t.RowsAffected() != int64(len(rows)) {
		return fmt.Errorf("mismatch: got %d rows affected; expected %d", t.RowsAffected(), len(rows))
	}

	return nil
}

func (t Table) IDs(ids []string) interface{} {
	switch t.IDCol.Type {
	case "integer", "bigint":
		return sqlutil.IntArray(intIDs(ids))
	case "uuid":
		return sqlutil.UUIDArray(ids)
	}
	return sqlutil.StringArray(ids)
}

type syncData struct {
	t        Table
	toInsert []syncRow
	toUpdate []syncRow
	toDelete []string
}

type syncRow struct {
	table string
	id    string
	data  json.RawMessage
}

func (rt *rowTracker) fetch(ctx context.Context, table Table, srcTx pgx.Tx, ids []string) (*syncData, error) {
	rows, err := srcTx.Query(ctx, table.SelectRowsQuery(), table.IDs(ids))
	if errors.Is(err, pgx.ErrNoRows) {
		return &syncData{toDelete: ids}, nil
	}
	defer rows.Close()
	if err != nil {
		return nil, fmt.Errorf("fetch rows: %w", err)
	}

	sd := syncData{t: table}
	existsInOld := make(map[string]struct{})
	for rows.Next() {
		var id string
		var data []byte
		err = rows.Scan(&id, &data)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		existsInOld[id] = struct{}{}
		if rt.Exists(table.Name, id) {
			sd.toUpdate = append(sd.toUpdate, syncRow{table.Name, id, data})
		} else {
			rt.Insert(table.Name, id)
			sd.toInsert = append(sd.toInsert, syncRow{table.Name, id, data})
		}
	}

	for _, id := range ids {
		if _, ok := existsInOld[id]; ok {
			continue
		}
		if !rt.Exists(table.Name, id) {
			continue
		}
		rt.Delete(table.Name, id)
		sd.toDelete = append(sd.toDelete, id)
	}

	return &sd, nil
}

func intIDs(ids []string) []int {
	var ints []int
	for _, id := range ids {
		i, err := strconv.Atoi(id)
		if err != nil {
			panic(err)
		}
		ints = append(ints, i)
	}
	return ints
}
