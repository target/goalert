package cleanupmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
)

var _ processinglock.Setupable = &DB{}

const QueueName = "cleanup-manager"

func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.CleanupAlerts))

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 2})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}

	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return AlertArgs{}, &river.InsertOpts{
					Queue: QueueName,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	return nil
}
