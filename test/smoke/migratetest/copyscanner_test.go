package migratetest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyScanner(t *testing.T) {
	const data = `
COPY "switchover_log" FROM stdin WITH (FORMAT csv, HEADER MATCH, ENCODING utf8);
id,timestamp,data
\.

COPY "switchover_state" FROM stdin WITH (FORMAT csv, HEADER MATCH, ENCODING utf8);
ok,current_state,db_id
t,idle,8df0bb48-404b-454f-96de-d851dbec0670
\.


COPY "config_limits" FROM stdin WITH (FORMAT csv, HEADER MATCH, ENCODING utf8);
id,max
notification_rules_per_user,15
contact_methods_per_user,10
\.

`

	s := NewCopyScanner(strings.NewReader(data))

	require.True(t, s.Scan())
	table := s.Table()
	require.Equal(t, "switchover_log", table.Name)
	require.Equal(t, []string{"id", "data", "timestamp"}, table.Columns)
	require.Empty(t, table.Rows)

	require.True(t, s.Scan())
	table = s.Table()
	require.Equal(t, "switchover_state", table.Name)
	require.Equal(t, []string{"current_state", "db_id", "ok"}, table.Columns)
	require.Equal(t, [][]string{
		{"idle", "8df0bb48-404b-454f-96de-d851dbec0670", "t"},
	}, table.Rows)

	require.True(t, s.Scan())
	table = s.Table()
	require.Equal(t, "config_limits", table.Name)
	require.Equal(t, []string{"id", "max"}, table.Columns)
	require.Equal(t, [][]string{ // sorted
		{"contact_methods_per_user", "10"},
		{"notification_rules_per_user", "15"},
	}, table.Rows)

	require.False(t, s.Scan())
	require.NoError(t, s.Err())
}
