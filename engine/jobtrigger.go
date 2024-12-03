package engine

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
)

// runAllPeriodicJobs schedules all periodic jobs immediately, and waits for them to complete.
func (p *Engine) runAllPeriodicJobs(ctx context.Context) error {
	var jobs []river.InsertManyParams
	for _, fn := range p.periodicJobs {
		args, opts := fn()
		jobs = append(jobs, river.InsertManyParams{Args: args, InsertOpts: opts})
	}
	res, err := p.cfg.River.InsertMany(ctx, jobs)
	if err != nil {
		return err
	}

	t := time.Tick(1 * time.Second)
	for _, j := range res {
		for {
			job, err := p.cfg.River.JobGet(ctx, j.Job.ID)
			if err != nil {
				return fmt.Errorf("job get: %w", err)
			}
			if !slices.Contains([]rivertype.JobState{rivertype.JobStateAvailable, rivertype.JobStateRunning, rivertype.JobStateScheduled, rivertype.JobStatePending}, job.State) {
				break // job is done
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t:
			}
		}
	}
	return nil
}
