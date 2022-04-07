package swo

import (
	"context"
	_ "embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swogrp"
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
	swogrp.Progressf(ctx, "scanning tables...")

	var cRow struct {
		TableName string
		Column
	}

	tables := make(map[string]*Table)
	_, err := conn.QueryFunc(ctx, columnListQuery, nil,
		[]interface{}{&cRow.TableName, &cRow.Column.Name, &cRow.Column.Type, &cRow.Column.Ord},
		func(pgx.QueryFuncRow) error {
			if tables[cRow.TableName] == nil {
				tables[cRow.TableName] = &Table{Name: cRow.TableName, deps: make(map[string]struct{})}
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
			tables[fRow.SrcName].deps[fRow.DstName] = struct{}{}

			return nil
		})
	if err != nil {
		return nil, err
	}

	var tableList []*Table
	for _, t := range tables {
		sort.Slice(t.Columns, func(i, j int) bool {
			return t.Columns[i].Ord < t.Columns[j].Ord
		})
		tableList = append(tableList, t)
	}

	// sort tables by name
	sort.Slice(tableList, func(i, j int) bool {
		return tableList[i].Name < tableList[j].Name
	})

	remove := func(i int) *Table {
		t := tableList[i]
		tableList = append(tableList[:i], tableList[i+1:]...)

		// delete table name from all deps
		for _, t2 := range tableList {
			delete(t2.deps, t.Name)
		}

		return t
	}
	next := func() *Table {
		for i, t := range tableList {
			if len(t.deps) == 0 {
				return remove(i)
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

func (c Column) IsInteger() bool {
	switch c.Type {
	case "integer", "bigint":
		return true
	}
	return false
}
