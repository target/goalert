package swoinfo

import (
	"context"
	_ "embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/target/goalert/swo/swodb"
)

// ScanTables scans the database for tables returning them in insert-safe-order,
// meaning the first table returned will not have any foreign keys to other tables.
//
// Tables with migrate-only data, or those used by switchover code will be omitted.
func ScanTables(ctx context.Context, conn *pgx.Conn) ([]Table, error) {
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
		switch cRow.ColTableName {
		case "engine_processing_versions", "gorp_migrations":
			// skip migrate-only tables
			continue
		case "switchover_state", "switchover_log", "change_log":
			// skip SWO tables
			continue
		}

		if tables[cRow.ColTableName] == nil {
			tables[cRow.ColTableName] = &Table{name: cRow.ColTableName, deps: make(map[string]struct{})}
		}

		tables[cRow.ColTableName].cols = append(tables[cRow.ColTableName].cols, column(cRow))
		if cRow.ColColumnName == "id" {
			tables[cRow.ColTableName].id = column(cRow)
		}
	}

	for _, t := range tables {
		if t.id.ColColumnName == "" {
			return nil, fmt.Errorf("table %s has no id column", t.name)
		}
	}

	for _, fRow := range refs {
		tables[fRow.SrcRelname].deps[fRow.DstRelname] = struct{}{}
	}

	var tableList []*Table
	for _, t := range tables {
		sort.Slice(t.cols, func(i, j int) bool {
			return t.cols[i].ColOrdinalPosition < t.cols[j].ColOrdinalPosition
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
