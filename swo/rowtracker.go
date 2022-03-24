package swo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
)

type rowTracker struct {
	tables []Table
	rowIDs map[string]map[string]struct{}
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

			rt.Insert(table.Name, id)
		}
	}

	return rt, nil
}

func (rt *rowTracker) Insert(table, id string)      { rt.rowIDs[table][id] = struct{}{} }
func (rt *rowTracker) Delete(table, id string)      { delete(rt.rowIDs[table], id) }
func (rt *rowTracker) Exists(table, id string) bool { _, ok := rt.rowIDs[table][id]; return ok }
