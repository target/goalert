package swoinfo

import "github.com/target/goalert/swo/swodb"

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
