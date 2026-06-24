package statusmgr

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
)

const (
	QueueName           = "status-manager"
	PriorityLookForWork = 2
	PriorityCleanup     = 3
	PriorityProcess     = 4
)

var _ processinglock.Setupable = &DB{}

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.cleanup))
	river.AddWorker(args.Workers, river.WorkFunc(db.processSubscription))
	river.AddWorker(args.Workers, river.WorkFunc(db.lookForWork))

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
					Priority: PriorityLookForWork,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return CleanupArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityCleanup,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	return nil
}
