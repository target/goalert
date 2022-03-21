package swo

import (
	"context"
	_ "embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v4"
)

type Column struct {
	Name string
	Type string
	Ord  int
}

var (
	//go:embed scantables_column_list.sql
	columnListQuery string

	//go:embed scantables_fkey_refs.sql
	fkeyRefsQuery string
)

// ScanTables scans the database for tables, their columns, and dependencies.
func ScanTables(ctx context.Context, conn *pgx.Conn) ([]Table, error) {
	var cRow struct {
		TableName string
		Column
	}

	tables := make(map[string]*Table)
	_, err := conn.QueryFunc(ctx, columnListQuery, nil,
		[]interface{}{&cRow.TableName, &cRow.Column.Name, &cRow.Column.Type, &cRow.Column.Ord},
		func(pgx.QueryFuncRow) error {
			if tables[cRow.TableName] == nil {
				tables[cRow.TableName] = &Table{Name: cRow.TableName, deps: make(map[string]*Table)}
			}
			tables[cRow.TableName].Columns = append(tables[cRow.TableName].Columns, cRow.Column)
			if cRow.Column.Name == "id" {
				tables[cRow.TableName].IDCol = cRow.Column
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("scanning table columns: %w", err)
	}

	var fRow struct {
		SrcName string
		DstName string
	}
	_, err = conn.QueryFunc(ctx, fkeyRefsQuery, nil, []interface{}{&fRow.SrcName, &fRow.DstName},
		func(pgx.QueryFuncRow) error {
			tables[fRow.SrcName].deps[fRow.DstName] = tables[fRow.DstName]

			return nil
		})
	if err != nil {
		return nil, err
	}

	// resolve/flatten dependencies
	var tableList []Table
	for _, t := range tables {
		tableList = append(tableList, *t)
		for t.flattenDeps() > 0 {
		}

		if _, ok := t.deps[t.Name]; ok {
			return nil, fmt.Errorf("circular non-deferrable dependency detected: %s", t.Name)
		}
	}

	// sort columns by ordinal
	for _, t := range tableList {
		sort.Slice(t.Columns, func(i, j int) bool {
			return t.Columns[i].Ord < t.Columns[j].Ord
		})
	}

	sort.Slice(tableList, func(i, j int) bool {
		if tableList[i].DependsOn(tableList[j].Name) {
			return false
		}
		if tableList[j].DependsOn(tableList[i].Name) {
			return true
		}

		return tableList[i].Name < tableList[j].Name
	})

	return tableList, nil
}

func (c Column) IsInteger() bool {
	switch c.Type {
	case "integer", "bigint":
		return true
	}
	return false
}
