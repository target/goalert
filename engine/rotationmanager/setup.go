package rotationmanager

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/event"
	"github.com/target/goalert/schedule/rotation"
)

const (
	QueueName           = "rotation-manager"
	PriorityLookForWork = 2
	PriorityCleanup     = 3
	PriorityProcess     = 4
)

var _ processinglock.Setupable = &DB{}

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.updateRotation))

	event.RegisterJobSource(args.EventBus, func(data rotation.Update) (river.JobArgs, *river.InsertOpts) {
		return UpdateArgs{RotationID: data.ID}, &river.InsertOpts{
			Queue:    QueueName,
			Priority: PriorityLookForWork,
		}
	})

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 5})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}

	return nil
}
