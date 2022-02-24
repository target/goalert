package smoketest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/target/goalert/smoketest/harness"
	"github.com/target/goalert/switchover/dbsync"
)

func TestDBSyncTables(t *testing.T) {
	t.Parallel()

	h := harness.NewHarness(t, "", "")
	defer h.Close()

	_, err := dbsync.Tables(context.Background(), h.App().DB())
	assert.NoError(t, err)
}
