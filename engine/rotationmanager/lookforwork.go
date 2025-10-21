package rotationmanager

import (
	"context"

	"github.com/riverqueue/river"
)

type LookForWorkArgs struct{}

func (LookForWorkArgs) Kind() string { return "rotation-manager-lfw" }

// lookForWork will schedule jobs for rotations in the entity_updates table.
func (db *DB) lookForWork(ctx context.Context, j *river.Job[LookForWorkArgs]) error {
	// No-op, this is handled by DB trigger directly now
	// We leave this here so any stale jobs get cleaned up
	return nil
}
