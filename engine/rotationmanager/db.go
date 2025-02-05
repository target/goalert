package rotationmanager

import (
	"context"
	"database/sql"

	"github.com/riverqueue/river"
	"github.com/target/goalert/engine/processinglock"
	"github.com/target/goalert/util"
)

// DB manages rotations in Postgres.
type DB struct {
	lock *processinglock.Lock

	currentTime *sql.Stmt

	lockPart   *sql.Stmt
	rotate     *sql.Stmt
	rotateData *sql.Stmt

	riverDBSQL *river.Client[*sql.Tx]
}

// Name returns the name of the module.
func (db *DB) Name() string { return "Engine.RotationManager" }

// NewDB will create a new DB, preparing all statements necessary.
func NewDB(ctx context.Context, db *sql.DB, riverDBSQL *river.Client[*sql.Tx]) (*DB, error) {
	lock, err := processinglock.NewLock(ctx, db, processinglock.Config{
		Type:    processinglock.TypeRotation,
		Version: 2,
	})
	if err != nil {
		return nil, err
	}
	p := &util.Prepare{Ctx: ctx, DB: db}

	return &DB{
		lock: lock,

		riverDBSQL: riverDBSQL,

		currentTime: p.P(`select now()`),
		lockPart:    p.P(`lock rotation_participants, rotation_state in exclusive mode`),
		rotate: p.P(`
			update rotation_state
			set
				shift_start = now(),
				rotation_participant_id = (select id from rotation_participants where rotation_id = $1 and position = $2),
				version = 2
			where rotation_id = $1
		`),
		rotateData: p.P(`
			select
				rot.id,
				rot."type",
				rot.start_time,
				rot.shift_length,
				rot.time_zone,
				state.shift_start,
				state."position",
				rot.participant_count,
				state.version
			from rotations rot
			join rotation_state state on state.rotation_id = rot.id
			where $1 or state.rotation_id = $2
			for update skip locked
		`),
	}, p.Err
}
