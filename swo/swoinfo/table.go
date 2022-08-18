package swoinfo

import (
	"fmt"
	"strings"

	"github.com/target/goalert/swo/swodb"
	"github.com/target/goalert/util/sqlutil"
)

type Table struct {
	name string
	deps map[string]struct{}
	cols []column
	id   column
}
type column swodb.InformationSchemaColumn

func (t Table) Name() string { return t.name }

func (t Table) IDType() string { return t.id.DataType }

func (t Table) Columns() []string {
	var cols []string
	for _, c := range t.cols {
		cols = append(cols, c.ColumnName)
	}
	return cols
}

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
