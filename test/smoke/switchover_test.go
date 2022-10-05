package smoke

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/switchover/dbsync"
	"github.com/target/goalert/test/smoke/harness"
)

// TestDBSyncTables ensures the latest state of the database is compatible with the dbsync package.
func TestDBSyncTables(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "")
	defer h.Close()

	_, err := dbsync.Tables(context.Background(), h.App().DB())
	assert.NoError(t, err)
}
