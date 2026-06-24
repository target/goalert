package schedulemanager

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/target/goalert/engine/processinglock"
)

// DB will manage schedules and schedule rules in Postgres.
type DB struct {
	lock *processinglock.Lock

	migrateSchedIDs []uuid.UUID
	migrateMap      map[uuid.UUID]uuid.UUID
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.ScheduleManager" }

// NewDB will create a new DB instance, preparing all statements.
func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeSchedule,
		Version: 3,
	})
	if err != nil {
		return nil, err
	}

	return &DB{lock: lock}, nil
}
