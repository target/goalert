package swo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v4"
)

type rowTracker struct {
	tables []Table
	rowIDs map[string]map[string]struct{}

	stagedInserts []stagedID
	stagedDeletes []stagedID
}
type stagedID struct {
	table string
	id    string
}

// ReadRowIDs reads the row IDs for all tables in the next-db to distinguish
// between those that need an INSERT vs UPDATE.
func (e *Execute) ReadRowIDs(ctx context.Context) {
	if e.err != nil {
		return
	}
	e.Progressf(ctx, "recording next DB row IDs")
	e.rowIDs = make(map[string]map[string]struct{})

	for _, table := range e.tables {
		if table.SkipSync() {
			continue
		}
		e.rowIDs[table.Name] = make(map[string]struct{})
		rows, err := e.nextDBConn.Query(ctx, fmt.Sprintf("SELECT id::text FROM %s", table.QuotedName()))
		if err != nil {
			e.err = fmt.Errorf("read row ids for %s: %w", table.Name, err)
			return
		}

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				e.err = fmt.Errorf("read row ids for %s: scan: %w", table.Name, err)
				return
			}

			e._Insert(table.Name, id)
		}
	}
}

func (e *Execute) Insert(table, id string) {
	e.stagedInserts = append(e.stagedInserts, stagedID{table, id})
}

func (e *Execute) Delete(table, id string) {
	e.stagedDeletes = append(e.stagedDeletes, stagedID{table, id})
}
func (e *Execute) _Insert(table, id string) { e.rowIDs[table][id] = struct{}{} }
func (e *Execute) _Delete(table, id string) { delete(e.rowIDs[table], id) }
func (e *Execute) Rollback() {
	e.stagedDeletes = nil
	e.stagedInserts = nil
}

func (e *Execute) Commit() {
	for _, staged := range e.stagedInserts {
		e._Insert(staged.table, staged.id)
	}
	e.stagedInserts = nil

	for _, staged := range e.stagedDeletes {
		e._Delete(staged.table, staged.id)
	}
	e.stagedDeletes = nil
}

func (e *Execute) Exists(table, id string) bool { _, ok := e.rowIDs[table][id]; return ok }

func (e *Execute) queueChanges(b *pgx.Batch, q string, rows []syncRow) error {
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

	b.Queue(q, data)

	return nil
}

func (e *Execute) fetchChanges(ctx context.Context, table Table, srcTx pgxQueryer, ids []string) (*syncData, error) {
	if len(ids) == 0 {
		return &syncData{}, nil
	}

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
		if e.Exists(table.Name, id) {
			sd.toUpdate = append(sd.toUpdate, syncRow{table.Name, id, data})
		} else {
			e.Insert(table.Name, id)
			sd.toInsert = append(sd.toInsert, syncRow{table.Name, id, data})
		}
	}

	for _, id := range ids {
		if _, ok := existsInOld[id]; ok {
			continue
		}
		if !e.Exists(table.Name, id) {
			continue
		}
		e.Delete(table.Name, id)
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
