package swo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
)

func (m *Manager) SendProposal() (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *Manager) ProposalIsValid() (bool, error) {
	return false, nil
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
		m.Progressf(ctx, "scanning tables...")
		tables, err := ScanTables(ctx, oldConn)
		if err != nil {
			return fmt.Errorf("scan tables: %w", err)
		}

		m.Progressf(ctx, "enabling change log")
		err = EnableChangeLog(ctx, tables, oldConn)
		if err != nil {
			return fmt.Errorf("enable change log: %w", err)
		}

		m.Progressf(ctx, "disabling triggers")
		err = DisableTriggers(ctx, tables, newConn)
		if err != nil {
			return fmt.Errorf("disable triggers: %w", err)
		}

		m.Progressf(ctx, "performing initial sync")
		err = m.InitialSync(ctx, oldConn, newConn)
		if err != nil {
			return fmt.Errorf("initial sync: %w", err)
		}

		m.Progressf(ctx, "recording new DB state")
		rt, err := newRowTracker(ctx, tables, newConn)
		if err != nil {
			return fmt.Errorf("read row IDs: %w", err)
		}

		var lastNone bool
		for ctx.Err() == nil {
			// sync in a loop until DB is up-to-date
			n, pend, err := LoopSync(ctx, rt, oldConn, newConn)
			if pend > 0 {
				lastNone = false
				m.Progressf(ctx, "sync: %d rows pending", pend)
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
					m.Progressf(ctx, "sync: waiting for changes")
				}
				time.Sleep(time.Second)
			} else {
				lastNone = false
				m.Progressf(ctx, "sync: %d rows replicated", n)
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

	n, err := syncChangeLog(ctx, rt, srcTx, dstTx)
	if err != nil {
		return 0, n, fmt.Errorf("sync change log: %w", err)
	}

	err = dstTx.Commit(ctx)
	if err != nil {
		return 0, n, fmt.Errorf("commit dst: %w", err)
	}

	err = srcTx.Commit(ctx)
	if err != nil {
		return n, 0, fmt.Errorf("commit src: %w", err)
	}

	return n, 0, nil
}

func FinalSync(ctx context.Context, oldConn, newConn *pgx.Conn) error {
	return nil
}

func syncTx(ctx context.Context, srcConn, dstConn *pgx.Conn) (src, dst pgx.Tx, err error) {
	srcTx, err := srcConn.BeginTx(ctx, pgx.TxOptions{
		AccessMode:     pgx.ReadWrite,
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

func syncChangeLog(ctx context.Context, rt *rowTracker, srcTx, dstTx pgx.Tx) (int, error) {
	type rowID struct {
		table string
		id    string
	}

	var r rowID
	changes := make(map[rowID]struct{})
	rowIDs := make(map[string][]string)
	_, err := srcTx.QueryFunc(ctx, "delete from change_log returning table_name, row_id", nil, []interface{}{&r.table, &r.id}, func(pgx.QueryFuncRow) error {
		if _, ok := changes[r]; ok {
			return nil
		}
		changes[r] = struct{}{}
		rowIDs[r.table] = append(rowIDs[r.table], r.id)

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("fetch changes: %w", err)
	}
	if len(changes) == 0 {
		return 0, nil
	}

	// defer all constraints
	_, err = dstTx.Exec(ctx, "SET CONSTRAINTS ALL DEFERRED")
	if err != nil {
		return len(changes), fmt.Errorf("defer constraints: %w", err)
	}

	type pendingDelete struct {
		query string
		idArg interface{}
	}
	var deletes []pendingDelete

	// go in insert order for fetching updates/inserts, note deleted rows
	for _, table := range rt.tables {
		if len(rowIDs[table.Name]) == 0 {
			continue
		}

		sd, err := rt.fetch(ctx, table, srcTx, rowIDs[table.Name])
		if err != nil {
			return len(changes), fmt.Errorf("fetch changed rows: %w", err)
		}
		if len(sd.toDelete) > 0 {
			deletes = append(deletes, pendingDelete{table.DeleteRowsQuery(), table.IDs(sd.toDelete)})
		}

		err = rt.apply(ctx, dstTx, table.UpdateRowsQuery(), sd.toUpdate)
		if err != nil {
			return len(changes), fmt.Errorf("apply updates: %w", err)
		}

		err = rt.apply(ctx, dstTx, table.InsertRowsQuery(), sd.toInsert)
		if err != nil {
			return len(changes), fmt.Errorf("apply inserts: %w", err)
		}
	}

	// handle pendingDeletes in reverse table order
	for i := len(deletes) - 1; i >= 0; i-- {
		_, err = dstTx.Exec(ctx, deletes[i].query, deletes[i].idArg)
		if err != nil {
			return len(changes), fmt.Errorf("delete rows: %w", err)
		}
	}

	return len(changes), nil
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
	_, err = dstTx.Exec(ctx, q, data)
	if err != nil {
		return fmt.Errorf("exec: %w", err)
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
