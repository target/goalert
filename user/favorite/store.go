package favorite

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// Store allows the lookup and management of Favorites.
type Store struct {
	db *sql.DB

	insert  *sql.Stmt
	delete  *sql.Stmt
	findAll *sql.Stmt
}

// NewStore will create a DB backend from a sql.DB. An error will be returned if statements fail to prepare.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
		db: db,
		insert: p.P(`
			INSERT INTO user_favorites (
				user_id, tgt_service_id,
				tgt_schedule_id,
				tgt_rotation_id, 
				tgt_escalation_policy_id, 
				tgt_user_id
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
		`),
		delete: p.P(`
			DELETE FROM user_favorites
			WHERE
				user_id = $1 AND
				tgt_service_id = $2 OR
				tgt_schedule_id = $3 OR
				tgt_rotation_id = $4 OR
				tgt_escalation_policy_id = $5 OR
				tgt_user_id = $6
		`),
		findAll: p.P(`
			SELECT
				tgt_service_id,
				tgt_schedule_id,
				tgt_rotation_id, 
				tgt_escalation_policy_id, 
				tgt_user_id
			FROM user_favorites
			WHERE user_id = $1
				AND (
					(tgt_service_id NOTNULL AND $2) OR
					(tgt_schedule_id NOTNULL AND $3) OR
					(tgt_rotation_id NOTNULL AND $4) OR
					(tgt_escalation_policy_id NOTNULL AND $5) OR
					(tgt_user_id NOTNULL AND $6)
				)
		`),
	}, p.Err
}

// Set will store the target as a favorite of the given user. Must be authorized as System or the same user.
// It is safe to call multiple times.
func (s *Store) Set(ctx context.Context, userID string, tgt assignment.Target) error {
	return s.SetTx(ctx, nil, userID, tgt)
}

// SetTx will store the target as a favorite of the given user. Must be authorized as System or the same user.
// It is safe to call multiple times.
func (s *Store) SetTx(ctx context.Context, tx *sql.Tx, userID string, tgt assignment.Target) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.MatchUser(userID))
	if err != nil {
		return err
	}
	err = validate.Many(
		validate.UUID("TargetID", tgt.TargetID()),
		validate.UUID("UserID", userID),
		validate.OneOf("TargetType", tgt.TargetType(), assignment.TargetTypeService,
			assignment.TargetTypeSchedule, assignment.TargetTypeRotation, assignment.TargetTypeEscalationPolicy, assignment.TargetTypeUser),
	)
	if err != nil {
		return err
	}
	stmt := s.insert
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	var serviceID, scheduleID, rotationID, epID, usrID sql.NullString
	switch tgt.TargetType() {
	case assignment.TargetTypeService:
		serviceID.Valid = true
		serviceID.String = tgt.TargetID()
	case assignment.TargetTypeSchedule:
		scheduleID.Valid = true
		scheduleID.String = tgt.TargetID()
	case assignment.TargetTypeRotation:
		rotationID.Valid = true
		rotationID.String = tgt.TargetID()
	case assignment.TargetTypeEscalationPolicy:
		epID.Valid = true
		epID.String = tgt.TargetID()
	case assignment.TargetTypeUser:
		usrID.Valid = true
		usrID.String = tgt.TargetID()
	}
	_, err = stmt.ExecContext(ctx, userID, serviceID, scheduleID, rotationID, epID, usrID)
	if err != nil {
		return errors.Wrap(err, "set favorite")
	}

	return nil
}

// Unset will remove the target as a favorite of the given user. Must be authorized as System or the same user.
// It is safe to call multiple times.
func (s *Store) Unset(ctx context.Context, userID string, tgt assignment.Target) error {
	err := permission.LimitCheckAny(ctx, permission.System, permission.MatchUser(userID))
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.UUID("TargetID", tgt.TargetID()),
		validate.UUID("UserID", userID),
		validate.OneOf("TargetType", tgt.TargetType(), assignment.TargetTypeService,
			assignment.TargetTypeSchedule, assignment.TargetTypeRotation, assignment.TargetTypeEscalationPolicy, assignment.TargetTypeUser),
	)
	if err != nil {
		return err
	}
	var serviceID, scheduleID, rotationID, epID, usrID sql.NullString
	switch tgt.TargetType() {
	case assignment.TargetTypeService:
		serviceID.Valid = true
		serviceID.String = tgt.TargetID()
	case assignment.TargetTypeSchedule:
		scheduleID.Valid = true
		scheduleID.String = tgt.TargetID()
	case assignment.TargetTypeRotation:
		rotationID.Valid = true
		rotationID.String = tgt.TargetID()
	case assignment.TargetTypeEscalationPolicy:
		epID.Valid = true
		epID.String = tgt.TargetID()
	case assignment.TargetTypeUser:
		usrID.Valid = true
		usrID.String = tgt.TargetID()
	}
	_, err = s.delete.ExecContext(ctx, userID, serviceID, scheduleID, rotationID, epID, usrID)
	if errors.Is(err, sql.ErrNoRows) {
		// ignoring since it is safe to unset favorite (with retries)
		err = nil
	}
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) FindAll(ctx context.Context, userID string, filter []assignment.TargetType) ([]assignment.Target, error) {
	err := permission.LimitCheckAny(ctx, permission.System, permission.MatchUser(userID))
	if err != nil {
		return nil, err
	}

	err = validate.Many(
		validate.UUID("UserID", userID),
		validate.Range("Filter", len(filter), 0, 50),
	)
	if err != nil {
		return nil, err
	}

	var allowServices, allowSchedules, allowRotations, allowEscalationPolicies, allowUsers bool
	if len(filter) == 0 {
		allowServices = true
	} else {
		for _, f := range filter {
			switch f {
			case assignment.TargetTypeService:
				allowServices = true
			case assignment.TargetTypeSchedule:
				allowSchedules = true
			case assignment.TargetTypeRotation:
				allowRotations = true
			case assignment.TargetTypeEscalationPolicy:
				allowEscalationPolicies = true
			case assignment.TargetTypeUser:
				allowUsers = true
			}
		}
	}

	rows, err := s.findAll.QueryContext(ctx, userID, allowServices, allowSchedules, allowRotations, allowEscalationPolicies, allowUsers)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var targets []assignment.Target

	for rows.Next() {
		var svc, sched, rot, escpolicy, usr sql.NullString
		err = rows.Scan(&svc, &sched, &rot, &escpolicy, &usr)
		if err != nil {
			return nil, err
		}
		switch {
		case svc.Valid:
			targets = append(targets, assignment.ServiceTarget(svc.String))
		case sched.Valid:
			targets = append(targets, assignment.ScheduleTarget(sched.String))
		case rot.Valid:
			targets = append(targets, assignment.RotationTarget(rot.String))
		case escpolicy.Valid:
			targets = append(targets, assignment.EscalationPolicyTarget(escpolicy.String))
		case usr.Valid:
			targets = append(targets, assignment.UserTarget(usr.String))
		}
	}
	return targets, nil
}
