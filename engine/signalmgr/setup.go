package signalmgr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/event"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/integrationkey/uik"
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

	event.RegisterJobSource(args.EventBus, func(data uik.EventNewSignals) (river.JobArgs, *river.InsertOpts) {
		return SchedMsgsArgs{
				ServiceID: uuid.NullUUID{Valid: true, UUID: data.ServiceID},
			}, &river.InsertOpts{
				Queue:       QueueName,
				ScheduledAt: time.Now().Add(time.Second),
				Priority:    PriorityScheduleService, // lower priority than the catch-all job
				UniqueOpts: river.UniqueOpts{
					ByArgs: true,
				},
			}
	})

	err := args.River.Queues().Add(QueueName, river.QueueConfig{MaxWorkers: 3})
	if err != nil {
		return fmt.Errorf("add queue: %w", err)
	}
	args.River.PeriodicJobs().AddMany([]*river.PeriodicJob{
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Hour),
			func() (river.JobArgs, *river.InsertOpts) {
				return MaintArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityMaintCleanup,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
		river.NewPeriodicJob(
			river.PeriodicInterval(time.Minute),
			func() (river.JobArgs, *river.InsertOpts) {
				return SchedMsgsArgs{}, &river.InsertOpts{
					Queue:    QueueName,
					Priority: PriorityScheduleAll,
				}
			},
			&river.PeriodicJobOpts{RunOnStart: true},
		),
	})

	return nil
}
