package dbsync

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx"
	"github.com/pkg/errors"
)

var (
	ignoreSyncTables = []string{
		"switchover_state",
		"engine_processing_versions",
		"gorp_migrations",
	}

	ignoreTriggerTables = append([]string{"change_log"}, ignoreSyncTables...)
)

type Table struct {
	Name    string
	Columns []Column
	IDCol   Column

	DependsOn   map[string]bool
	DependantOf map[string]bool
}
type Column struct {
	Name string
	Type string
	Ord  int
}

func contains(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

func Tables(ctx context.Context, db *sql.DB) ([]Table, error) {
	rows, err := db.QueryContext(ctx, `
		select col.table_name, col.column_name, col.data_type, col.ordinal_position
		from information_schema.columns col
		join information_schema.tables t on
			t.table_catalog = col.table_catalog and
			t.table_schema = col.table_schema and
			t.table_name = col.table_name and
			t.table_type = 'BASE TABLE'
		where col.table_catalog = current_database() and col.table_schema = 'public'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	t := make(map[string]*Table)
	for rows.Next() {
		var col Column
		var name string
		err = rows.Scan(&name, &col.Name, &col.Type, &col.Ord)
		if err != nil {
			return nil, err
		}
		if contains(ignoreSyncTables, name) {
			continue
		}
		tbl := t[name]
		if tbl == nil {
			tbl = &Table{Name: name, DependsOn: make(map[string]bool), DependantOf: make(map[string]bool)}
			t[name] = tbl
		}
		tbl.Columns = append(tbl.Columns, col)
		if col.Name == "id" {
			tbl.IDCol = col
		}
	}

	rows, err = db.QueryContext(ctx, `
		select src.relname, ref.relname
		from pg_catalog.pg_constraint con
		join pg_namespace ns on ns.nspname = 'public' and ns.oid = con.connamespace
		join pg_class src on src.oid = con.conrelid
		join pg_class ref on ref.oid = con.confrelid
		where con.contype = 'f' and not con.condeferrable
	`)
	if err != nil {
		return nil, errors.Wrap(err, "fetch non-deferrable dependencies")
	}
	defer rows.Close()
	for rows.Next() {
		var srcName, refName string
		err = rows.Scan(&srcName, &refName)
		if err != nil {
			return nil, errors.Wrap(err, "scan non-deferrable dependency")
		}
		t[srcName].DependsOn[refName] = true
		t[refName].DependantOf[srcName] = true
		if t[refName].DependsOn[srcName] {
			return nil, errors.Errorf("circular non-deferrable dependency between '%s' and '%s'", srcName, refName)
		}
	}

	var isRecursiveDep func(a, b string) bool
	isRecursiveDep = func(a, b string) bool {
		// if 'a' depends on 'b'
		if t[a].DependsOn[b] {
			return true
		}

		// if a dep of 'a' depends on 'b'
		for dep := range t[a].DependsOn {
			if isRecursiveDep(dep, b) {
				return true
			}
		}

		return false
	}
	var recursiveDependants func(Table) []Table
	recursiveDependants = func(tbl Table) []Table {
		var tables []Table
		for name := range tbl.DependantOf {
			tables = append(tables, *t[name])
			tables = append(tables, recursiveDependants(*t[name])...)
		}
		return tables
	}

	tables := make([]Table, 0, len(t))
	for _, tbl := range t {
		sort.Slice(tbl.Columns, func(i, j int) bool { return tbl.Columns[i].Ord < tbl.Columns[j].Ord })
		tables = append(tables, *tbl)
	}

	// sort by name
	sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })

	// sort by deps
	depOrder := make([]Table, 0, len(tables))
	deps := make(map[string]bool)
	for len(depOrder) < len(tables) {
	tableLoop:
		for _, t := range tables {
			if deps[t.Name] {
				continue
			}
			for depName := range t.DependsOn {
				if !deps[depName] {
					continue tableLoop
				}
			}
			deps[t.Name] = true
			depOrder = append(depOrder, t)
		}
	}

	return depOrder, nil
}

func (c Column) IsInteger() bool {
	switch c.Type {
	case "integer", "bigint":
		return true
	}
	return false
}
func (t Table) SafeName() string {
	return pgx.Identifier{t.Name}.Sanitize()
}
func (t Table) ColumnNames() []string {
	colNames := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		colNames[i] = col.Name
	}
	return colNames
}

func (t Table) FetchOneRow() string {
	return fmt.Sprintf(`select * from %s where id = cast($1 as %s)`, t.SafeName(), t.IDCol.Type)
}
func (t Table) DeleteOneRow() string {
	return fmt.Sprintf(`delete from %s where id = cast($1 as %s)`, t.SafeName(), t.IDCol.Type)
}
func (t Table) InsertOneRow() string {
	return fmt.Sprintf(`
		insert into %s
		select * from
		json_populate_record(null::%s, $1)
		as data
	`,
		t.SafeName(),
		t.SafeName(),
	)
}

func (t Table) UpdateOneRow() string {
	cols := make([]string, 0, len(t.Columns))
	for _, col := range t.Columns {
		if col.Name == "id" {
			continue
		}
		cols = append(cols, fmt.Sprintf(`%s = data.%s`, pgx.Identifier{col.Name}.Sanitize(), pgx.Identifier{col.Name}.Sanitize()))
	}

	return fmt.Sprintf(`
		update %s dst
		set %s
		from (select * from json_populate_record(null::%s, $2)) as data
		where dst.id = $1
	`,
		t.SafeName(),
		strings.Join(cols, ", "),
		t.SafeName(),
	)
}
