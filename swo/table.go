package swo

import (
	"fmt"
	"strings"

	"github.com/target/goalert/util/sqlutil"
)

// Table describes a database table for a switchover operation.
type Table struct {
	Name    string
	Columns []Column
	IDCol   Column

	deps map[string]*Table
}

func (t *Table) DependsOn(name string) bool {
	return t.deps[name] != nil
}

func (t *Table) flattenDeps() int {
	var n int
	for _, tbl := range t.deps {
		for name, dep := range tbl.deps {
			if _, ok := t.deps[name]; ok {
				continue
			}
			t.deps[name] = dep
			n++
		}
	}

	return n
}

// SkipSync returns true if the table should not be synced or instrumented with triggers.
//
// This could be because the data comes from migration or is stateful/related
// to the switchover.
func (t Table) SkipSync() bool {
	switch t.Name {
	case "switchover_state", "switchover_log", "engine_processing_versions", "gorp_migrations", "change_log":
		return true
	}

	return false
}

func (t Table) QuotedName() string {
	return sqlutil.QuoteID(t.Name)
}

func (t Table) QuotedChangeTriggerName() string {
	return sqlutil.QuoteID(fmt.Sprintf("zz_99_change_log_%s", t.Name))
}

func (t Table) QuotedLockTriggerName() string {
	return sqlutil.QuoteID(fmt.Sprintf("!_change_log_%s", t.Name))
}

func (t Table) ColumnNames() []string {
	colNames := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		colNames[i] = col.Name
	}
	return colNames
}

func (t Table) SelectOneRowQuery() string {
	return fmt.Sprintf(`select * from %s where id = cast($1 as %s)`, t.QuotedName(), t.IDCol.Type)
}

func (t Table) DeleteOneRowQuery() string {
	return fmt.Sprintf(`delete from %s where id = cast($1 as %s)`, t.QuotedName(), t.IDCol.Type)
}

func (t Table) InsertOneRowQuery() string {
	return fmt.Sprintf(`
		insert into %s
		select * from
		json_populate_record(null::%s, $1)
		as data
	`,
		t.QuotedName(),
		t.QuotedName(),
	)
}

func (t Table) UpdateOneRowQuery() string {
	cols := make([]string, 0, len(t.Columns))
	for _, col := range t.Columns {
		if col.Name == "id" {
			continue
		}
		cols = append(cols, fmt.Sprintf(`%s = data.%s`, sqlutil.QuoteID(col.Name), sqlutil.QuoteID(col.Name)))
	}

	return fmt.Sprintf(`
		update %s dst
		set %s
		from (select * from json_populate_record(null::%s, $2)) as data
		where dst.id = $1
	`,
		t.QuotedName(),
		strings.Join(cols, ", "),
		t.QuotedName(),
	)
}
