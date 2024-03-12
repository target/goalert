package rotation

import (
	"context"
	"database/sql"
	"sort"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// ErrNoState is returned when there is no state information available for a rotation.
var ErrNoState = errors.New("no state available")

type Store struct {
	db *sql.DB

	createRotation        *sql.Stmt
	updateRotation        *sql.Stmt
	findRotation          *sql.Stmt
	findRotationForUpdate *sql.Stmt
	deleteRotation        *sql.Stmt
	findMany              *sql.Stmt

	findAllParticipants *sql.Stmt
	addParticipant      *sql.Stmt
	findParticipant     *sql.Stmt

	state    *sql.Stmt
	rmState  *sql.Stmt
	lockPart *sql.Stmt

	deleteParticipants      *sql.Stmt
	updateParticipantUserID *sql.Stmt
	setActiveIndex          *sql.Stmt

	findPartCount *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db: db,

		lockPart: p.P(`lock rotation_participants, rotation_state in exclusive mode`),

		createRotation: p.P(`INSERT INTO rotations (id, name, description, type, start_time, shift_length, time_zone) VALUES ($1, $2, $3, $4, $5, $6, $7)`),
		updateRotation: p.P(`
			WITH set_shift_start AS (
				UPDATE rotation_state
				SET shift_start = now()
				WHERE rotation_id = $1
			)
			UPDATE rotations SET name = $2, description = $3, type = $4, start_time = $5, shift_length = $6, time_zone = $7 WHERE id = $1
		`),
		findRotation: p.P(`
			SELECT 
				r.id, 
				r.name, 
				r.description, 
				r.type, 
				r.start_time, 
				r.shift_length, 
				r.time_zone, 
				fav IS DISTINCT FROM NULL 
			FROM rotations r 
			LEFT JOIN user_favorites fav ON fav.tgt_rotation_id = r.id 
			AND fav.user_id = $2 
			WHERE r.id = $1
		`),
		findRotationForUpdate: p.P(`SELECT id, name, description, type, start_time, shift_length, time_zone FROM rotations WHERE id = $1 FOR UPDATE`),
		deleteRotation:        p.P(`DELETE FROM rotations WHERE id = ANY($1)`),

		findMany: p.P(`
			SELECT 
				r.id, 
				r.name, 
				r.description, 
				r.type, 
				r.start_time, 
				r.shift_length, 
				r.time_zone,
				fav IS DISTINCT FROM NULL 
			FROM rotations r 
			LEFT JOIN user_favorites fav ON fav.tgt_rotation_id = r.id 
			AND fav.user_id = $2 
			WHERE r.id = ANY($1)
		`),

		addParticipant: p.P(`
			INSERT INTO rotation_participants (id, rotation_id, position, user_id)
			VALUES (
				$1,
				$2,
				0,
				$3
			)
			RETURNING position
		`),

		findAllParticipants: p.P(`SELECT id, rotation_id, position, user_id FROM rotation_participants WHERE rotation_id = $1 ORDER BY position`),

		findParticipant: p.P(`SELECT rotation_id, position, user_id FROM rotation_participants WHERE id = $1`),

		state: p.P(`
			SELECT
				position,
				rotation_participant_id,
				shift_start
			FROM rotation_state
			WHERE rotation_id = $1
		`),
		rmState: p.P(`
			DELETE FROM rotation_state WHERE rotation_id = $1
		`),

		deleteParticipants: p.P(`
			DELETE FROM rotation_participants WHERE id = ANY($1)
		`),

		updateParticipantUserID: p.P(`
			UPDATE rotation_participants SET user_id = $2 WHERE id = $1
		`),

		setActiveIndex: p.P(`
			UPDATE rotation_state SET rotation_participant_id = (SELECT id FROM rotation_participants WHERE rotation_id = $1 AND position = $2),
			position = $2
			WHERE rotation_id = $1
		`),
		findPartCount: p.P(`SELECT participant_count FROM rotations WHERE id = $1`),
	}, p.Err
}

func (s *Store) State(ctx context.Context, id string) (*State, error) {
	return s.StateTx(ctx, nil, id)
}

func (s *Store) StateTx(ctx context.Context, tx *sql.Tx, id string) (*State, error) {
	err := validate.UUID("RotationID", id)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	stmt := s.state
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	row := stmt.QueryRowContext(ctx, id)
	var st State
	var part sql.NullString
	err = row.Scan(&st.Position, &part, &st.ShiftStart)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoState
	}
	if err != nil {
		return nil, errors.Wrap(err, "query rotation state")
	}
	st.ParticipantID = part.String
	st.RotationID = id

	return &st, nil
}

func (s *Store) CreateRotationTx(ctx context.Context, tx *sql.Tx, r *Rotation) (*Rotation, error) {
	n, err := r.Normalize()
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := s.createRotation
	if tx != nil {
		stmt = tx.Stmt(stmt)
	}

	n.ID = uuid.New().String()

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.Type, n.Start, n.ShiftLength, n.Start.Location().String())
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Store) UpdateRotationTx(ctx context.Context, tx *sql.Tx, r *Rotation) error {
	err := validate.UUID("RotationID", r.ID)
	if err != nil {
		return err
	}
	n, err := r.Normalize()
	if err != nil {
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}

	stmt := s.updateRotation
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.Type, n.Start, n.ShiftLength, n.Start.Location().String())
	return err
}

func (s *Store) FindMany(ctx context.Context, ids []string) ([]Rotation, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.ManyUUID("RotationID", ids, 200)
	if err != nil {
		return nil, err
	}

	rows, err := s.findMany.QueryContext(ctx, sqlutil.UUIDArray(ids), permission.NullUserUUID(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var r Rotation
	var tz string
	result := make([]Rotation, 0, len(ids))
	for rows.Next() {
		err = rows.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.Start, &r.ShiftLength, &tz, &r.isUserFavorite)
		if err != nil {
			return nil, err
		}
		loc, err := util.LoadLocation(tz)
		if err != nil {
			return nil, err
		}
		r.Start = r.Start.In(loc)
		result = append(result, r)
	}

	return result, nil
}

func (s *Store) FindRotation(ctx context.Context, id string) (*Rotation, error) {
	err := validate.UUID("RotationID", id)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	row := s.findRotation.QueryRowContext(ctx, id, permission.NullUserUUID(ctx))
	var r Rotation
	var tz string
	err = row.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.Start, &r.ShiftLength, &tz, &r.isUserFavorite)
	if err != nil {
		return nil, err
	}
	loc, err := util.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	r.Start = r.Start.In(loc)
	return &r, nil
}

func (s *Store) FindParticipantCount(ctx context.Context, id string) (int, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return -1, err
	}

	err = validate.UUID("RotationID", id)
	if err != nil {
		return -1, err
	}

	row := s.findPartCount.QueryRowContext(ctx, id)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (s *Store) FindRotationForUpdateTx(ctx context.Context, tx *sql.Tx, rotationID string) (*Rotation, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("RotationID", rotationID)
	if err != nil {
		return nil, err
	}

	stmt := s.findRotationForUpdate
	if tx != nil {
		_, err = tx.StmtContext(ctx, s.lockPart).ExecContext(ctx)
		if err != nil {
			return nil, err
		}

		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, rotationID)
	var r Rotation
	var tz string
	err = row.Scan(&r.ID, &r.Name, &r.Description, &r.Type, &r.Start, &r.ShiftLength, &tz)
	if err != nil {
		return nil, err
	}
	loc, err := util.LoadLocation(tz)
	if err != nil {
		return nil, err
	}
	r.Start = r.Start.In(loc)
	return &r, nil
}

func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("RotationID", ids, 50)
	if err != nil {
		return err
	}

	return s.withTxLock(ctx, tx, func(tx *sql.Tx) error {
		_, err := tx.StmtContext(ctx, s.deleteRotation).ExecContext(ctx, sqlutil.UUIDArray(ids))
		return err
	})
}

func (s *Store) FindAllParticipantsTx(ctx context.Context, tx *sql.Tx, rotationID string) ([]Participant, error) {
	err := validate.UUID("RotationID", rotationID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := s.findAllParticipants
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	rows, err := stmt.QueryContext(ctx, rotationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var p Participant
	var userID sql.NullString
	var res []Participant
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.RotationID, &p.Position, &userID)
		if err != nil {
			return nil, err
		}
		if userID.Valid {
			p.Target = assignment.UserTarget(userID.String)
		} else {
			p.Target = nil
		}
		res = append(res, p)
	}

	sort.Slice(res, func(i, j int) bool { return res[i].Position < res[j].Position })

	return res, nil
}

func (s *Store) FindAllParticipants(ctx context.Context, rotationID string) ([]Participant, error) {
	return s.FindAllParticipantsTx(ctx, nil, rotationID)
}

func (s *Store) withTxLock(ctx context.Context, tx *sql.Tx, f func(*sql.Tx) error) error {
	if tx != nil {
		_, err := tx.StmtContext(ctx, s.lockPart).ExecContext(ctx)
		if err != nil {
			return err
		}

		return f(tx)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer sqlutil.Rollback(ctx, "rotation: lock participants", tx)

	err = s.withTxLock(ctx, tx, f)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) SetActiveIndexTx(ctx context.Context, tx *sql.Tx, rotID string, position int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("RotationID", rotID),
	)
	if err != nil {
		return err
	}

	stmt := s.setActiveIndex
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, rotID, position)
	if e := sqlutil.MapError(err); e != nil && e.Code == "23502" && e.ColumnName == "rotation_participant_id" {
		// 23502 is not_null_violation
		// https://www.postgresql.org/docs/9.6/errcodes-appendix.html
		// We are checking to see if there is no participant for that position before returning a validation error
		return validation.NewFieldError("ActiveUserIndex", "invalid index for rotation")
	}
	return err
}

func (s *Store) FindParticipant(ctx context.Context, id string) (*Participant, error) {
	err := validate.UUID("RotationParticipantID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	row := s.findParticipant.QueryRowContext(ctx, id)
	var p Participant
	p.ID = id
	var userID sql.NullString
	err = row.Scan(&p.RotationID, &p.Position, &userID)
	if userID.Valid {
		p.Target = assignment.UserTarget(userID.String)
	}

	return &p, err
}

func (s *Store) AddRotationUsersTx(ctx context.Context, tx *sql.Tx, rotationID string, userIDs []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.ManyUUID("UserIDs", userIDs, 50)
	if err != nil {
		return err
	}

	err = s.withTxLock(ctx, tx, func(tx *sql.Tx) error {
		stmt := tx.StmtContext(ctx, s.addParticipant)
		for _, userID := range userIDs {
			_, err = stmt.ExecContext(ctx, uuid.New().String(), rotationID, userID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *Store) DeleteRotationParticipantsTx(ctx context.Context, tx *sql.Tx, partIDs []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.ManyUUID("ParticipantIDs", partIDs, 50)
	if err != nil {
		return err
	}

	return s.withTxLock(ctx, tx, func(tx *sql.Tx) error {
		_, err := tx.StmtContext(ctx, s.deleteParticipants).ExecContext(ctx, sqlutil.UUIDArray(partIDs))
		return err
	})
}

func (s *Store) UpdateParticipantUserIDTx(ctx context.Context, tx *sql.Tx, partID, userID string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("ParticipantID", partID),
		validate.UUID("UserID", userID),
	)
	if err != nil {
		return err
	}

	return s.withTxLock(ctx, tx, func(tx *sql.Tx) error {
		_, err := tx.StmtContext(ctx, s.updateParticipantUserID).ExecContext(ctx, partID, userID)
		return err
	})
}

func (s *Store) DeleteStateTx(ctx context.Context, tx *sql.Tx, rotationID string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.UUID("RotationID", rotationID)
	if err != nil {
		return err
	}

	return s.withTxLock(ctx, tx, func(tx *sql.Tx) error {
		_, err := tx.StmtContext(ctx, s.rmState).ExecContext(ctx, rotationID)
		return err
	})
}
