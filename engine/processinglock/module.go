package processinglock

import (
	"context"
	"database/sql"

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

// SetupArgs is a struct that contains the arguments for the setup function.
type SetupArgs struct {
	DB           *sql.DB
	River        *river.Client[pgx.Tx]
	Workers      *river.Workers
	ConfigSource config.Source
}

// Setupable is an interface for types that can be set up using the job queue system.
type Setupable interface {
	Module
	Setup(context.Context, SetupArgs) error
}
