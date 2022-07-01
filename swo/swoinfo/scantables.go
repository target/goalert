package swoinfo

import (
	"context"
	_ "embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swogrp"
)

// ScanTables scans the database for tables returning them in insert-safe-order,
// meaning the first table returned will not have any foreign keys to other tables.
//
// Tables with migrate-only data, or those used by switchover code will be omitted.
func ScanTables(ctx context.Context, conn *pgx.Conn) ([]Table, error) {
	swogrp.Progressf(ctx, "scanning tables...")

	columns, err := swodb.New(conn).TableColumns(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan table columns: %w", err)
	}

	refs, err := swodb.New(conn).ForeignKeyRefs(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan foreign keys: %w", err)
	}

	tables := make(map[string]*Table)
	for _, cRow := range columns {
		switch cRow.TableName {
		case "engine_processing_versions", "gorp_migrations":
			// skip migrate-only tables
			continue
		case "switchover_state", "switchover_log", "change_log":
			// skip SWO tables
			continue
		}

		if tables[cRow.TableName] == nil {
			tables[cRow.TableName] = &Table{name: cRow.TableName, deps: make(map[string]struct{})}
		}

		tables[cRow.TableName].cols = append(tables[cRow.TableName].cols, column(cRow))
		if cRow.ColumnName == "id" {
			tables[cRow.TableName].id = column(cRow)
		}
	}

	for _, t := range tables {
		if t.id.ColumnName == "" {
			return nil, fmt.Errorf("table %s has no id column", t.name)
		}
	}

	for _, fRow := range refs {
		tables[fRow.SrcRelname].deps[fRow.DstRelname] = struct{}{}
	}

	var tableList []*Table
	for _, t := range tables {
		sort.Slice(t.cols, func(i, j int) bool {
			return t.cols[i].OrdinalPosition < t.cols[j].OrdinalPosition
		})
		tableList = append(tableList, t)
	}

	// sort tables by name
	sort.Slice(tableList, func(i, j int) bool {
		return tableList[i].name < tableList[j].name
	})

	// take the next table, remove it from other dependency lists
	pick := func(i int) *Table {
		t := tableList[i]
		tableList = append(tableList[:i], tableList[i+1:]...)

		// delete table name from all deps
		for _, t2 := range tableList {
			delete(t2.deps, t.name)
		}

		return t
	}

	// get the next table to pick (zero dependencies)
	next := func() *Table {
		for i, t := range tableList {
			if len(t.deps) == 0 {
				return pick(i)
			}
		}

		return nil
	}

	var result []Table
	for {
		t := next()
		if t == nil {
			break
		}
		result = append(result, *t)
	}

	return result, nil
}
