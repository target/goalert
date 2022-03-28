package swo

import (
	"context"
	"fmt"

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

func newRowTracker(ctx context.Context, tables []Table, newConn *pgx.Conn) (*rowTracker, error) {
	rt := &rowTracker{
		tables: tables,
		rowIDs: make(map[string]map[string]struct{}),
	}

	for _, table := range tables {
		if table.SkipSync() {
			continue
		}
		rt.rowIDs[table.Name] = make(map[string]struct{})
		rows, err := newConn.Query(ctx, fmt.Sprintf("SELECT id::text FROM %s", table.QuotedName()))
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return nil, err
			}

			rt._Insert(table.Name, id)
		}
	}

	return rt, nil
}

func (rt *rowTracker) Insert(table, id string) {
	rt.stagedInserts = append(rt.stagedInserts, stagedID{table, id})
}

func (rt *rowTracker) Delete(table, id string) {
	rt.stagedDeletes = append(rt.stagedDeletes, stagedID{table, id})
}
func (rt *rowTracker) _Insert(table, id string) { rt.rowIDs[table][id] = struct{}{} }
func (rt *rowTracker) _Delete(table, id string) { delete(rt.rowIDs[table], id) }
func (rt *rowTracker) Rollback() {
	rt.stagedDeletes = nil
	rt.stagedInserts = nil
}

func (rt *rowTracker) Commit() {
	for _, staged := range rt.stagedInserts {
		rt._Insert(staged.table, staged.id)
	}
	rt.stagedInserts = nil

	for _, staged := range rt.stagedDeletes {
		rt._Delete(staged.table, staged.id)
	}
	rt.stagedDeletes = nil
}

func (rt *rowTracker) Exists(table, id string) bool { _, ok := rt.rowIDs[table][id]; return ok }
