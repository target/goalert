package signalmgr

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/gadb"
)

const (
	QueueName               = "engine-signal-mgr"
	PriorityScheduleAll     = 2
	PriorityMaintCleanup    = 3
	PriorityScheduleService = 4
)

var _ processinglock.Setupable = &DB{}

type MaintArgs struct{}

func (MaintArgs) Kind() string { return "signal-manager-cleanup" }

type SchedMsgsArgs struct {
	ServiceID uuid.NullUUID
}

func (SchedMsgsArgs) Kind() string { return "signal-manager-schedule-outgoing-messages" }

func (db *DB) Setup(ctx context.Context, args processinglock.SetupArgs) error {
	river.AddWorker(args.Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[MaintArgs]) error {
		return db.lock.WithTxShared(ctx, func(ctx context.Context, tx *sql.Tx) error {
			return gadb.New(tx).SignalMgrDeleteStale(ctx)
		})
	}))

	river.AddWorker(args.Workers, river.WorkFunc(func(ctx context.Context, j *river.Job[SchedMsgsArgs]) error {
		return db.scheduleMessages(ctx, j.Args.ServiceID)
	}))

	args.AddQueue(QueueName, 3)
	args.AddPeriodicJob(time.Hour, func() (river.JobArgs, *river.InsertOpts) {
		return MaintArgs{}, &river.InsertOpts{
			Queue:    QueueName,
			Priority: PriorityMaintCleanup,
		}
	})
	args.AddPeriodicJob(time.Minute, func() (river.JobArgs, *river.InsertOpts) {
		return SchedMsgsArgs{}, &river.InsertOpts{
			Queue:    QueueName,
			Priority: PriorityScheduleAll,
		}
	})

	return nil
}
