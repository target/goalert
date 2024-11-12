package signalmgr

import (
	"context"
	"database/sql"
	"time"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/gadb"
)

const QueueName string = "engine-signal-mgr"

var _ processinglock.Setupable = &DB{}

type MaintArgs struct{}

func (MaintArgs) Kind() string { return "signal-mgr-maint" }

type SchedMsgsArgs struct{}

func (SchedMsgsArgs) Kind() string { return "signal-mgr-sched-msgs" }

func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[MaintArgs]) error {
		return db.lock.WithTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
			return gadb.New(tx).SignalMgrDeleteStale(ctx)
		})
	}))

	river.AddWorker(args.Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[SchedMsgsArgs]) error {
		return db.scheduleMessages(ctx)
	}))

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 1})
	if err != nil {
		return err
	}

	jobs := args.River.PeriodicJobs()
	jobs.Add(river.NewPeriodicJob(
		river.PeriodicInterval(time.Hour),
		func() (river.JobArgs, *river.InsertOpts) {
			return MaintArgs{}, &river.InsertOpts{
				Queue: QueueName,
			}
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	))
	jobs.Add(river.NewPeriodicJob(
		river.PeriodicInterval(time.Minute),
		func() (river.JobArgs, *river.InsertOpts) {
			return SchedMsgsArgs{}, &river.InsertOpts{
				Queue: QueueName,
			}
		},
		&river.PeriodicJobOpts{RunOnStart: true},
	))

	return nil
}
