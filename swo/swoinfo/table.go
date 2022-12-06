package swoinfo

import (
	"fmt"
	"strings"

	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/util/sqlutil"
)

// Table represents a table in the database.
type Table struct {
	name string
	deps map[string]struct{}
	cols []column
	id   column
}
type column swodb.TableColumnsRow

// Name returns the name of the table.
func (t Table) Name() string { return t.name }

// IDType returns the type of the ID column.
func (t Table) IDType() string { return t.id.ColDataType }

// Columns returns the names of the columns in the table.
func (t Table) Columns() []string {
	var cols []string
	for _, c := range t.cols {
		cols = append(cols, c.ColColumnName)
	}
	return cols
}

// InsesrtJSONRowsQuery returns a query that can be used to insert or upsert rows from the given JSON data.
func (t Table) InsertJSONRowsQuery(upsert bool) string {
	query := fmt.Sprintf("insert into %s select * from json_populate_recordset(null::%s, $1)", sqlutil.QuoteID(t.Name()), sqlutil.QuoteID(t.Name()))
	if !upsert {
		return query
	}

	sets := make([]string, 0, len(t.cols))
	for _, col := range t.Columns() {
		if col == "id" {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = excluded.%s", sqlutil.QuoteID(col), sqlutil.QuoteID(col)))
	}

	return fmt.Sprintf("%s on conflict (id) do update set %s where %s.id = excluded.id", query, strings.Join(sets, ", "), sqlutil.QuoteID(t.Name()))
}
