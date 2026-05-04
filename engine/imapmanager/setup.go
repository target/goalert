package imapmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
)

var _ processinglock.Setupable = &DB{}

const QueueName = "imap-manager"

const (
	PriorityPoll    = 1
	PriorityCleanup = 2
)

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.PollIMAP))
	river.AddWorker(args.Workers, river.WorkFunc(db.CleanupProcessedMessages))

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 3})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}

	// Add periodic polling job every 1 minute (per-service intervals checked at runtime)
	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(1*time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return PollArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityPoll,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	// Add daily cleanup job to remove old processed messages
	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(24*time.Hour),
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
