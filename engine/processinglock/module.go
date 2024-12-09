package processinglock

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/target/goalert/event"
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

// SetupArgs is a struct that contains the arguments for the setup function.
type SetupArgs struct {
	DB       *sql.DB
	Workers  *river.Workers
	EventBus *event.Bus
	River    *river.Client[pgx.Tx]
}

// Setupable is an interface for types that can be set up using the job queue system.
type Setupable interface {
	Module

	// Setup is called to configure the processing lock system. Any workers, queues, and periodic jobs should be added here.
	Setup(context.Context, SetupArgs) error
}
