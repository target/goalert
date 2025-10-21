package rotationmanager

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
)

const (
	QueueName         = "rotation-manager"
	PriorityScheduled = 1
	PriorityEvent     = 2
	PriorityLFW       = 4
)

var _ processinglock.Setupable = (*DB)(nil) // assert that DB implements processinglock.Setupable

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.updateRotation))
	river.AddWorker(args.Workers, river.WorkFunc(db.lookForWork))

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 5})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}

	return nil
}
