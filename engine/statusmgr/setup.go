package statusmgr

import (
	"context"
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

func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.cleanup))
	river.AddWorker(args.Workers, river.WorkFunc(db.processSubscription))
	river.AddWorker(args.Workers, river.WorkFunc(db.lookForWork))

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 5})
	if err != nil {
		return err
	}

	jobs := args.River.PeriodicJobs()
	jobs.Add(river.NewPeriodicJob(
		river.PeriodicInterval(time.Hour),
		func() (river.JobArgs, *river.InsertOpts) {
			return CleanupArgs{}, &river.InsertOpts{
				Queue:    QueueName,
				Priority: PriorityCleanup,
			}
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	))

	jobs.Add(river.NewPeriodicJob(
		river.PeriodicInterval(time.Second*5),
		func() (river.JobArgs, *river.InsertOpts) {
			return LookForWorkArgs{}, &river.InsertOpts{
				Queue:    QueueName,
				Priority: PriorityLookForWork,
			}
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	))

	return nil
}
