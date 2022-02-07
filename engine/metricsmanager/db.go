package metricsmanager

import (
	"context"
	"database/sql"

	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB handles updating metrics
type DB struct {
	db   *sql.DB
	lock *processinglock.Lock

	setTimeout       *sql.Stmt
	findNextAlertIDs *sql.Stmt

	findCurrentState 	*sql.Stmt
	findMaxAlertID		*sql.Stmt
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.MetricsManager" }

// NewDB creates a new DB.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Version: 1,
		Type:    processinglock.TypeMetrics,
	})
	if err != nil {
		return nil, err
	}

	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		db:   db,
		lock: lock,

		// Abort any cleanup operation that takes longer than 3 seconds
		// error will be logged.
		setTimeout: p.P(`SET LOCAL statement_timeout = 3000`),

		findNextAlertIDs: p.P(`select id from alerts limit 3000`),

		findCurrentState: p.P(`select state -> (select 'V' || version::text from engine_processing_versions where type_id = 'metrics') as state from engine_processing_versions where type_id = 'metrics'; `),

		findMaxAlertID: p.P(`select max(id) from alerts`),
	}, p.Err
}
