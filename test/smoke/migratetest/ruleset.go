package migratetest

import (
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/pgdump-lite"
)

// IgnoreRule is a rule to ignore differences in a snapshot.
type IgnoreRule struct {
	// MigrationName is the name of the migration this rule applies to.
	MigrationName string
	TableName     string
	ColumnName    string

	// ExtraRows will ignore extra/leftover rows in the table when migrating down.
	ExtraRows bool

	// MissingRows will ignore missing rows in the table when migrating down (e.g., table was dropped).
	MissingRows bool
}

// RuleSet is a set of IgnoreRules.
type RuleSet []IgnoreRule

type tableRules struct {
	AllowExtra    bool
	AllowMissing  bool
	IgnoreColumns []string
}

func (rs RuleSet) rulesForTable(t *testing.T, tableName string, isDown bool) tableRules {
	t.Helper()

	var r tableRules
	migrationName := strings.Split(t.Name(), "/")[1]
	for _, rule := range rs {
		if rule.MigrationName != "" && rule.MigrationName != migrationName {
			continue
		}
		if rule.TableName != tableName {
			continue
		}

		if isDown && rule.ExtraRows {
			r.AllowExtra = true
		}
		if isDown && rule.MissingRows {
			r.AllowMissing = true
		}
		if rule.ColumnName != "" {
			r.IgnoreColumns = append(r.IgnoreColumns, rule.ColumnName)
		}
	}

	return r
}

// RequireEqualDown will compare two snapshots and ignore any differences based on the rules in the RuleSet after a Down migration.
func (rs RuleSet) RequireEqualDown(t *testing.T, expected, actual *Snapshot) {
	t.Helper()

	require.Subset(t, names(actual.Schema.Extensions), names(expected.Schema.Extensions), "Extensions were removed that shouldn't have been") // extensions can be added, but don't need to be removed

	requireSameEntities(t, expected.Schema.Functions, actual.Schema.Functions, "Functions")
	requireSameEntities(t, expected.Schema.Sequences, actual.Schema.Sequences, "Sequences")
	requireSameEntities(t, expected.Schema.Tables, actual.Schema.Tables, "Tables")

	requireSameEntitiesWith(t, expected.Schema.Enums, actual.Schema.Enums, "Enums", func(t *testing.T, e, act pgdump.Enum) {
		t.Helper()

		require.Subsetf(t, act.Values, e.Values, "Enum values from %s were removed that shouldn't have been", e.Name)
	})

	requireSameEntitiesWith(t, expected.TableData, actual.TableData, "Table Data", rs.makeRequireTableDataMatch(true))
}

type stringNameable interface {
	String() string
	nameable
}

func requireSameEntities[T stringNameable](t *testing.T, expected, actual []T, typeName string) {
	t.Helper()

	requireSameEntitiesWith(t, expected, actual, typeName, func(t *testing.T, e, act T) {
		t.Helper()

		require.Equalf(t, e.String(), act.String(), "%s %s has wrong definition", typeName, e.EntityName())
	})
}

func requireSameEntitiesWith[T nameable](t *testing.T, expected, actual []T, typeName string, compare func(*testing.T, T, T)) {
	t.Helper()

	require.Equal(t, names(expected), names(actual), typeName+" mismatch")

	for _, e := range expected {
		act, ok := byName(actual, e.EntityName())
		if !ok {
			t.Fatalf("%s %s was removed and should not have been", typeName, e.EntityName())
		}

		compare(t, e, act)
	}
}

type nameable interface {
	EntityName() string
}

func byName[T nameable](items []T, name string) (value T, ok bool) {
	for _, i := range items {
		if i.EntityName() == name {
			return i, true
		}
	}
	return value, false
}

func names[T nameable](items []T) []string {
	var out []string
	for _, i := range items {
		out = append(out, i.EntityName())
	}
	return out
}

// AssertEqualUp will compare two snapshots and ignore any differences based on the rules in the RuleSet after the second Up migration.
func (rs RuleSet) RequireEqualUp(t *testing.T, expected, actual *Snapshot) {
	t.Helper()

	// schema should be identical once re-applied
	requireSameEntities(t, expected.Schema.Functions, actual.Schema.Functions, "Functions")
	requireSameEntities(t, expected.Schema.Sequences, actual.Schema.Sequences, "Sequences")
	requireSameEntities(t, expected.Schema.Tables, actual.Schema.Tables, "Tables")
	requireSameEntities(t, expected.Schema.Extensions, actual.Schema.Extensions, "Extensions")
	requireSameEntities(t, expected.Schema.Enums, actual.Schema.Enums, "Enums")

	requireSameEntitiesWith(t, expected.TableData, actual.TableData, "Table Data", rs.makeRequireTableDataMatch(false))
}

func (rs RuleSet) makeRequireTableDataMatch(isDown bool) func(*testing.T, TableSnapshot, TableSnapshot) {
	return func(t *testing.T, exp, act TableSnapshot) {
		t.Helper()

		// ensure we're comparing apples to apples
		require.Equalf(t, exp.Columns, act.Columns, "Table %s has wrong columns", exp.Name)

		rules := rs.rulesForTable(t, exp.Name, isDown)
		ignoreColIdx := make([]int, len(rules.IgnoreColumns))
		for i, col := range rules.IgnoreColumns {
			ignoreColIdx[i] = slices.Index(exp.Columns, col)
		}

		var hasErr bool
		if !rules.AllowMissing {
			// find any missing expected rows
			for _, row := range exp.Rows {
				if !containsRow(act.Rows, row, ignoreColIdx) {
					t.Errorf("Table %s missing row: %v", exp.Name, row)
					hasErr = true
				}
			}
		}

		if !rules.AllowExtra {
			// find any extra rows in actual
			for _, row := range act.Rows {
				if !containsRow(exp.Rows, row, ignoreColIdx) {
					t.Errorf("Table %s has extra row: %v", exp.Name, row)
					hasErr = true
				}
			}
		}

		if hasErr {
			t.FailNow()
		}
	}
}

func containsRow(rows [][]string, row []string, ignoreCols []int) bool {
	for _, r := range rows {
		if rowsMatch(r, row, ignoreCols) {
			return true
		}
	}
	return false
}

func rowsMatch(a, b []string, ignoreCols []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if slices.Contains(ignoreCols, i) {
			continue
		}
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
