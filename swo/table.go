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

	deps map[string]struct{}
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
		colNames[i] = col.ColumnName
	}
	return colNames
}

func (t Table) SelectRowsQuery() string {
	if t.IDCol.DataType == "USER-DEFINED" {
		return fmt.Sprintf(`select id::text, to_jsonb(row) from %s row where id::text = any($1)`, t.QuotedName())
	}
	return fmt.Sprintf(`select id::text, to_jsonb(row) from %s row where id = any($1)`, t.QuotedName())
}

func (t Table) DeleteRowsQuery() string {
	return fmt.Sprintf(`delete from %s where id = any($1)`, t.QuotedName())
}

func (t Table) InsertRowsQuery() string {
	return fmt.Sprintf(`
		insert into %s
		select * from
		json_populate_recordset(null::%s, $1)
	`, t.QuotedName(), t.QuotedName())
}

func (t Table) UpdateRowsQuery() string {
	var set strings.Builder
	for i, col := range t.Columns {
		if i > 0 {
			set.WriteString(", ")
		}
		set.WriteString(fmt.Sprintf("%s = data.%s", col.ColumnName, col.ColumnName))
	}

	return fmt.Sprintf(`
		update %s dst
		set %s
		from json_populate_recordset(null::%s, $1) as data
		where dst.id = data.id
	`, t.QuotedName(), set.String(), t.QuotedName())
}
