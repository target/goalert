package rotationmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/event"
	"github.com/target/goalert/schedule/rotation"
)

const (
	QueueName         = "rotation-manager"
	PriorityScheduled = 1
	PriorityEvent     = 2
	PriorityLFW       = 4
)

var _ processinglock.Setupable = &DB{}

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.updateRotation))
	river.AddWorker(args.Workers, river.WorkFunc(db.lookForWork))

	event.RegisterJobSource(args.EventBus, func(data rotation.Update) (river.JobArgs, *river.InsertOpts) {
		return UpdateArgs{RotationID: data.ID}, &river.InsertOpts{
			Queue:    QueueName,
			Priority: PriorityEvent,
		}
	})

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 5})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}

	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return LookForWorkArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityLFW,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	return nil
}
