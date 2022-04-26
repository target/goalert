package rule

import (
	"context"
	"database/sql"
	"errors"

	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
)

type ScheduleTriggerFunc func(string)
type Store struct {
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

func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
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

func (s *Store) FindScheduleID(ctx context.Context, ruleID string) (string, error) {
	err := validate.UUID("RuleID", ruleID)
	if err != nil {
		return "", err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return "", err
	}
	row := s.findScheduleID.QueryRowContext(ctx, ruleID)
	var schedID string
	err = row.Scan(&schedID)
	if err != nil {
		return "", err
	}
	return schedID, nil
}

func (s *Store) _Add(ctx context.Context, stmt *sql.Stmt, r *Rule) (*Rule, error) {
	n, err := r.Normalize()
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	n.ID = uuid.New().String()
	_, err = stmt.ExecContext(ctx, n.readFields()...)
	if err != nil {
		return nil, err
	}

	return n, nil
}

func (s *Store) Add(ctx context.Context, r *Rule) (*Rule, error) {
	r, err := s._Add(ctx, s.add, r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Store) CreateRuleTx(ctx context.Context, tx *sql.Tx, r *Rule) (*Rule, error) {
	if tx == nil {
		return s._Add(ctx, s.add, r)
	}
	return s._Add(ctx, tx.Stmt(s.add), r)
}

func (s *Store) FindByTargetTx(ctx context.Context, tx *sql.Tx, scheduleID string, target assignment.Target) ([]Rule, error) {
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

	stmt := s.findTgt
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
	if errors.Is(err, sql.ErrNoRows) {
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
func (s *Store) DeleteByTarget(ctx context.Context, scheduleID string, target assignment.Target) error {
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
	_, err = s.deleteAssignmentByTarget.ExecContext(ctx, scheduleID, tgtUser, tgtRot)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, ruleID string) error {
	return s.DeleteTx(ctx, nil, ruleID)
}
func (s *Store) DeleteTx(ctx context.Context, tx *sql.Tx, ruleID string) error {
	return s.DeleteManyTx(ctx, tx, []string{ruleID})
}
func (s *Store) DeleteManyTx(ctx context.Context, tx *sql.Tx, ruleIDs []string) error {
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
	stmt := s.delete
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, sqlutil.UUIDArray(ruleIDs))
	return err

}

func (s *Store) UpdateTx(ctx context.Context, tx *sql.Tx, r *Rule) error {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}
	n, err := r.Normalize()
	if err != nil {
		return err
	}

	f := n.readFields()

	stmt := s.update
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	_, err = stmt.ExecContext(ctx, f...)
	if err != nil {
		return err
	}
	return nil
}
func (s *Store) Update(ctx context.Context, r *Rule) error {
	return s.UpdateTx(ctx, nil, r)
}

func (s *Store) FindOne(ctx context.Context, ruleID string) (*Rule, error) {
	err := validate.UUID("RuleID", ruleID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	var r Rule
	err = r.scanFrom(s.findOne.QueryRowContext(ctx, ruleID))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// FindAllWithUsers works like FindAll but resolves rotations to the active user.
// This is reflected in the Target attribute.
// Rules pointing to inactive rotations (no participants) are omitted.
func (s *Store) FindAllWithUsers(ctx context.Context, scheduleID string) ([]Rule, error) {
	err := validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	rows, err := s.findAllUsers.QueryContext(ctx, scheduleID)
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
func (s *Store) FindAll(ctx context.Context, scheduleID string) ([]Rule, error) {
	return s.FindAllTx(ctx, nil, scheduleID)
}

func (s *Store) FindAllTx(ctx context.Context, tx *sql.Tx, scheduleID string) ([]Rule, error) {
	err := validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	stmt := s.findAll
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
