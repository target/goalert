package migratetest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/target/goalert/devtools/pgdump-lite"
)

type IgnoreRule struct {

	// MigrationName is the name of the migration this rule applies to.
	MigrationName string

	// TableName is the name of the table this rule applies to.
	TableName string

	// ColumnName is the name of the column this rule applies to (if applicable).
	ColumnName string

	// ExtraRows will ignore extra/leftover rows in the table when migrating down.
	ExtraRows bool

	// MissingRows will ignore missing rows in the table when migrating down (e.g., table was dropped).
	MissingRows bool
}

type RuleSet []IgnoreRule

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
}

func requireSameEntities[T nameable](t *testing.T, expected, actual []T, typeName string) {
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
	String() string
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
}
