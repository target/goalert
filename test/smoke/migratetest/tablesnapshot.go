package migratetest

import (
	"slices"
	"sort"
	"strings"
)

// TableSnapshot is a snapshot of a table's data.
type TableSnapshot struct {
	// Name is the name of the table.
	Name    string
	Columns []string
	Rows    [][]string
}

func (t TableSnapshot) EntityName() string { return t.Name }

type columnSort TableSnapshot

func (data *columnSort) Len() int { return len(data.Columns) }
func (data *columnSort) Less(i, j int) bool {
	// sort by column name, but prefer "id" as the first column
	if data.Columns[i] == "id" {
		return true
	}
	if data.Columns[j] == "id" {
		return false
	}

	return data.Columns[i] < data.Columns[j]
}
func (data *columnSort) Swap(i, j int) {
	data.Columns[i], data.Columns[j] = data.Columns[j], data.Columns[i]
	for _, row := range data.Rows {
		row[i], row[j] = row[j], row[i]
	}
}

// Sort sorts the columns and rows of the snapshot.
func (data *TableSnapshot) Sort() {
	sort.Sort((*columnSort)(data))

	// sort rows by first column, then second, etc.
	slices.SortFunc(data.Rows, func(a, b []string) int {
		for i := range a {
			if a[i] != b[i] {
				return strings.Compare(a[i], b[i])
			}
		}

		return 0
	})
}
