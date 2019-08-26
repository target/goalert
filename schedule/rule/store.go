package rule

import (
	"context"
	"database/sql"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	uuid "github.com/satori/go.uuid"
)

type Store interface {
	ReadStore
	Add(context.Context, *Rule) (*Rule, error)
	CreateRuleTx(context.Context, *sql.Tx, *Rule) (*Rule, error)
	Update(context.Context, *Rule) error
	UpdateTx(context.Context, *sql.Tx, *Rule) error
	Delete(context.Context, string) error
	DeleteTx(context.Context, *sql.Tx, string) error
	DeleteManyTx(context.Context, *sql.Tx, []string) error
	DeleteByTarget(ctx context.Context, scheduleID string, target assignment.Target) error
	FindByTargetTx(ctx context.Context, tx *sql.Tx, scheduleID string, target assignment.Target) ([]Rule, error)
}
type ReadStore interface {
	FindScheduleID(context.Context, string) (string, error)
	FindOne(context.Context, string) (*Rule, error)
	FindAll(ctx context.Context, scheduleID string) ([]Rule, error)
	FindAllTx(ctx context.Context, tx *sql.Tx, scheduleID string) ([]Rule, error)

	// FindAllWithUsers works like FindAll but resolves rotations to the active user.
	// This is reflected in the Target attribute.
	// Rules pointing to inactive rotations (no participants) are omitted.
	FindAllWithUsers(ctx context.Context, scheduleID string) ([]Rule, error)
}
type ScheduleTriggerFunc func(string)
type DB struct {
	db *sql.DB

	add     *sql.Stmt
	update  *sql.Stmt
	delete  *sql.Stmt
	findOne *sql.Stmt
	findAll *sql.Stmt
	findTgt *sql.Stmt

	deleteAssignmentByTarget *sql.Stmt

	findAllUsers *sql.Stmt

	findScheduleID *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &DB{
		db: db,

		findScheduleID: p.P(`
			select schedule_id
			from schedule_rules
			where id = $1
		`),
		add: p.P(`
			insert into schedule_rules (
				id,
				schedule_id,
				sunday,
				monday,
				tuesday,
				wednesday,
				thursday,
				friday,
				saturday,
				start_time,
				end_time,
				tgt_user_id,
				tgt_rotation_id
			) values ($1, $2, ($3::Bool[])[1], ($3::Bool[])[2], ($3::Bool[])[3], ($3::Bool[])[4], ($3::Bool[])[5], ($3::Bool[])[6], ($3::Bool[])[7], $4, $5, $6, $7)
		`),
		update: p.P(`
			update schedule_rules
			set
				schedule_id = $2,
				sunday = ($3::Bool[])[1],
				monday = ($3::Bool[])[2],
				tuesday = ($3::Bool[])[3],
				wednesday = ($3::Bool[])[4],
				thursday = ($3::Bool[])[5],
				friday = ($3::Bool[])[6],
				saturday = ($3::Bool[])[7],
				start_time = $4,
				end_time = $5,
				tgt_user_id = $6,
				tgt_rotation_id = $7
			where id = $1
		`),
		delete: p.P(`delete from schedule_rules where id = any($1)`),
		deleteAssignmentByTarget: p.P(`
			delete from schedule_rules
			where
				schedule_id = $1 and
				(tgt_user_id = $2 or
				tgt_rotation_id = $3)
		`),
		findOne: p.P(`
			select
				id,
				schedule_id,
				ARRAY[
					sunday,
					monday,
					tuesday,
					wednesday,
					thursday,
					friday,
					saturday
				],
				start_time,
				end_time,
				tgt_user_id,
				tgt_rotation_id
			from schedule_rules
			where id = $1
		`),

		findAll: p.P(`
			select
				id,
				schedule_id,
				ARRAY[
					sunday,
					monday,
					tuesday,
					wednesday,
					thursday,
					friday,
					saturday
				],
				start_time,
				end_time,
				tgt_user_id,
				tgt_rotation_id
			from schedule_rules
			where schedule_id = $1
			order by created_at, id
		`),
		findTgt: p.P(`
			select
				id,
				schedule_id,
				ARRAY[
					sunday,
					monday,
					tuesday,
					wednesday,
					thursday,
					friday,
					saturday
				],
				start_time,
				end_time,
				tgt_user_id,
				tgt_rotation_id
			from schedule_rules
			where schedule_id = $1 AND (tgt_user_id = $2 OR tgt_rotation_id = $3)
			order by created_at, id
		`),
		findAllUsers: p.P(`
			with rotation_users as (
				select
					s.rotation_id,
					p.user_id
				from rotation_state s
				join rotation_participants p on s.rotation_participant_id = p.id
			)
			select
				id,
				schedule_id,
				ARRAY[
					sunday,
					monday,
					tuesday,
					wednesday,
					thursday,
					friday,
					saturday
				],
				start_time,
				end_time,
				case when tgt_user_id is not null then
					tgt_user_id
				else
					rUser.user_id
				end,
				null
			from schedule_rules r
			left join rotation_users rUser on rUser.rotation_id = r.tgt_rotation_id
			where schedule_id = $1
			order by created_at, id
		`),
	}, p.Err
}

func (db *DB) FindScheduleID(ctx context.Context, ruleID string) (string, error) {
	err := validate.UUID("RuleID", ruleID)
	if err != nil {
		return "", err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return "", err
	}
	row := db.findScheduleID.QueryRowContext(ctx, ruleID)
	var schedID string
	err = row.Scan(&schedID)
	if err != nil {
		return "", err
	}
	return schedID, nil
}

func (db *DB) _Add(ctx context.Context, s *sql.Stmt, r *Rule) (*Rule, error) {
	n, err := r.Normalize()
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	n.ID = uuid.NewV4().String()
	_, err = s.ExecContext(ctx, n.readFields()...)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (db *DB) Add(ctx context.Context, r *Rule) (*Rule, error) {
	r, err := db._Add(ctx, db.add, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (db *DB) CreateRuleTx(ctx context.Context, tx *sql.Tx, r *Rule) (*Rule, error) {
	return db._Add(ctx, tx.Stmt(db.add), r)
}

func (db *DB) FindByTargetTx(ctx context.Context, tx *sql.Tx, scheduleID string, target assignment.Target) ([]Rule, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.Many(
		validate.UUID("ScheduleID", scheduleID),
		validate.OneOf("TargetType", target.TargetType(), assignment.TargetTypeUser, assignment.TargetTypeRotation),
		validate.UUID("TargetID", target.TargetID()),
	)
	if err != nil {
		return nil, err
	}

	stmt := db.findTgt
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	var tgtUser, tgtRot sql.NullString
	switch target.TargetType() {
	case assignment.TargetTypeUser:
		tgtUser.Valid = true
		tgtUser.String = target.TargetID()
	case assignment.TargetTypeRotation:
		tgtRot.Valid = true
		tgtRot.String = target.TargetID()
	}

	rows, err := stmt.QueryContext(ctx, scheduleID, tgtUser, tgtRot)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Rule
	var r Rule
	for rows.Next() {
		err = r.scanFrom(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}

// DeleteByTarget removes all rules for a schedule pointing to the specified target.
func (db *DB) DeleteByTarget(ctx context.Context, scheduleID string, target assignment.Target) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("ScheduleID", scheduleID),
		validate.OneOf("TargetType", target.TargetType(), assignment.TargetTypeUser, assignment.TargetTypeRotation),
		validate.UUID("TargetID", target.TargetID()),
	)
	if err != nil {
		return err
	}

	var tgtUser, tgtRot sql.NullString

	switch target.TargetType() {
	case assignment.TargetTypeUser:
		tgtUser.Valid = true
		tgtUser.String = target.TargetID()
	case assignment.TargetTypeRotation:
		tgtRot.Valid = true
		tgtRot.String = target.TargetID()
	}
	_, err = db.deleteAssignmentByTarget.ExecContext(ctx, scheduleID, tgtUser, tgtRot)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) Delete(ctx context.Context, ruleID string) error {
	return db.DeleteTx(ctx, nil, ruleID)
}
func (db *DB) DeleteTx(ctx context.Context, tx *sql.Tx, ruleID string) error {
	return db.DeleteManyTx(ctx, tx, []string{ruleID})
}
func (db *DB) DeleteManyTx(ctx context.Context, tx *sql.Tx, ruleIDs []string) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}
	if len(ruleIDs) == 0 {
		return nil
	}
	err = validate.ManyUUID("RuleIDs", ruleIDs, 50)
	if err != nil {
		return err
	}
	s := db.delete
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}

	_, err = s.ExecContext(ctx, sqlutil.UUIDArray(ruleIDs))
	return err

}

func (db *DB) UpdateTx(ctx context.Context, tx *sql.Tx, r *Rule) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}
	n, err := r.Normalize()
	if err != nil {
		return err
	}

	f := n.readFields()

	stmt := db.update
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	_, err = stmt.ExecContext(ctx, f...)
	if err != nil {
		return err
	}
	return nil
}
func (db *DB) Update(ctx context.Context, r *Rule) error {
	return db.UpdateTx(ctx, nil, r)
}

func (db *DB) FindOne(ctx context.Context, ruleID string) (*Rule, error) {
	err := validate.UUID("RuleID", ruleID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	var r Rule
	err = r.scanFrom(db.findOne.QueryRowContext(ctx, ruleID))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (db *DB) FindAllWithUsers(ctx context.Context, scheduleID string) ([]Rule, error) {
	err := validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	rows, err := db.findAllUsers.QueryContext(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Rule
	var r Rule
	for rows.Next() {
		err = r.scanFrom(rows)
		if err == errNoKnownTarget {
			err = nil
		}
		if err != nil {
			return nil, err
		}
		if r.Target == nil || r.Target.TargetType() != assignment.TargetTypeUser {
			continue
		}
		result = append(result, r)
	}

	return result, nil
}
func (db *DB) FindAll(ctx context.Context, scheduleID string) ([]Rule, error) {
	return db.FindAllTx(ctx, nil, scheduleID)
}

func (db *DB) FindAllTx(ctx context.Context, tx *sql.Tx, scheduleID string) ([]Rule, error) {
	err := validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	stmt := db.findAll
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	rows, err := stmt.QueryContext(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Rule
	var r Rule
	for rows.Next() {
		err = r.scanFrom(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}

	return result, nil
}
