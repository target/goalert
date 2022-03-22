package smoketest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/target/goalert/smoketest/harness"
	"github.com/target/goalert/swo"
)

// TestDBSyncTables ensures the latest state of the database is compatible with the dbsync package.
func TestDBSyncTables(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "")
	defer h.Close()

	c, err := h.App().DB().Conn(context.Background())
	require.NoError(t, err)
	defer c.Close()

	err = c.Raw(func(c interface{}) error {
		_, err := swo.ScanTables(context.Background(), c.(*stdlib.Conn).Conn())
		return err
	})
	assert.NoError(t, err)
}
