package swo

import (
	"context"
	_ "embed"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v4"
	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/swo/swogrp"
)

type Column swodb.InformationSchemaColumn

// ScanTables scans the database for tables, their columns, and dependencies.
func ScanTables(ctx context.Context, conn *pgx.Conn) ([]Table, error) {
	swogrp.Progressf(ctx, "scanning tables...")

	columns, err := swodb.New(conn).TableColumns(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan table columns: %w", err)
	}

	tables := make(map[string]*Table)
	for _, cRow := range columns {
		if tables[cRow.TableName] == nil {
			tables[cRow.TableName] = &Table{Name: cRow.TableName, deps: make(map[string]struct{})}
		}

		tables[cRow.TableName].Columns = append(tables[cRow.TableName].Columns, Column(cRow))
		if cRow.ColumnName == "id" {
			tables[cRow.TableName].IDCol = Column(cRow)
		}
	}

	refs, err := swodb.New(conn).ForeignKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("scan foreign keys: %w", err)
	}
	for _, fRow := range refs {
		tables[fRow.SrcRelname].deps[fRow.DstRelname] = struct{}{}
	}

	var tableList []*Table
	for _, t := range tables {
		sort.Slice(t.Columns, func(i, j int) bool {
			return t.Columns[i].OrdinalPosition < t.Columns[j].OrdinalPosition
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
