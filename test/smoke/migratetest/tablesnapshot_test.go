package migratetest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTableSnapshot_Sort(t *testing.T) {
	var data TableSnapshot
	data.Columns = []string{"value", "id", "name"}
	data.Rows = [][]string{
		{"1", "id_b", "foo"},
		{"2", "id_a", "baz"},
		{"3", "id_a", "bar"},
	}

	data.Sort()

	assert.Equal(t, []string{"id", "name", "value"}, data.Columns)
	assert.Equal(t, [][]string{
		{"id_a", "bar", "3"},
		{"id_a", "baz", "2"},
		{"id_b", "foo", "1"},
	}, data.Rows)
}
