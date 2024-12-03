package processinglock

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/config"
)

// Module is a processing lock module.
type Module interface {
	Name() string
}

// Updatable is an interface for types that can be updated.
type Updatable interface {
	Module
	UpdateAll(context.Context) error
}

func NewSetupArgs(river *river.Client[pgx.Tx], registerJobConstFn func(river.PeriodicJobConstructor)) SetupArgs {
	return SetupArgs{river: river, regJobFn: registerJobConstFn}
}

// SetupArgs is a struct that contains the arguments for the setup function.
type SetupArgs struct {
	DB           *sql.DB
	Workers      *river.Workers
	ConfigSource config.Source
	river        *river.Client[pgx.Tx]
	regJobFn     func(river.PeriodicJobConstructor)
}

// AddQueue configures a queue with the given name and maxWorkers.
func (a SetupArgs) AddQueue(name string, maxWorkers int) {
	err := a.river.Queues().Add(name, river.QueueConfig{MaxWorkers: maxWorkers})
	if err != nil {
		// Indicates invalid worker config or queue name (which should be static).
		panic(err)
	}
}

// AddPeriodicJob adds a periodic job to river, while registering it with the engine for manual triggering during tests.
func (a SetupArgs) AddPeriodicJob(dur time.Duration, fn river.PeriodicJobConstructor) {
	a.regJobFn(fn)
	a.river.PeriodicJobs().Add(river.NewPeriodicJob(
		river.PeriodicInterval(dur),
		fn,
		&river.PeriodicJobOpts{RunOnStart: true},
	))
}

// Setupable is an interface for types that can be set up using the job queue system.
type Setupable interface {
	Module

	// Setup is called to configure the processing lock system. Any workers, queues, and periodic jobs should be added here.
	Setup(context.Context, SetupArgs) error
}
