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

// Setup implements processinglock.Setupable.
func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(db.cleanup))
	river.AddWorker(args.Workers, river.WorkFunc(db.processSubscription))
	river.AddWorker(args.Workers, river.WorkFunc(db.lookForWork))

	args.AddQueue(QueueName, 5)
	args.AddPeriodicJob(time.Second*5, LookForWorkArgs{}, &river.InsertOpts{
		Queue:    QueueName,
		Priority: PriorityCleanup,
	})
	args.AddPeriodicJob(time.Hour, CleanupArgs{}, &river.InsertOpts{
		Queue:    QueueName,
		Priority: PriorityCleanup,
	})

	return nil
}
