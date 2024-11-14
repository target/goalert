package signalmgr

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// TriggerService will attempt to schedule messages for the given service immediately.
func TriggerService(ctx context.Context, r *river.Client[pgx.Tx], serviceID uuid.UUID) error {
	res, err := r.Insert(ctx, SchedMsgsArgs{
		ServiceID: uuid.NullUUID{Valid: true, UUID: serviceID},
	}, &river.InsertOpts{
		Queue:       QueueName,
		ScheduledAt: time.Now().Add(time.Second),
		Priority:    PriorityScheduleService, // lower priority than the catch-all job
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	})
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	if res.UniqueSkippedAsDuplicate {
		_, err = r.JobRetry(ctx, res.Job.ID)
		if err != nil {
			return fmt.Errorf("retry job: %w", err)
		}
	}

	return nil
}
