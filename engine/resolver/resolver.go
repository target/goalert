package resolver

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
)

type Resolver interface {
	AlertEPID(context.Context, int) (string, error)
	IsUserOnCall(ctx context.Context, userID string) (bool, error)
	OnCallByUser(ctx context.Context, userID string) ([]OnCallAssignment, error)
}

type OnCallAssignment struct {
	ServiceID    string `json:"service_id"`
	ServiceName  string `json:"service_name"`
	EPID         string `json:"escalation_policy_id"`
	EPName       string `json:"escalation_policy_name"`
	Level        int    `json:"escalation_policy_step_number"`
	RotationID   string `json:"rotation_id"`
	RotationName string `json:"rotation_name"`
	ScheduleID   string `json:"schedule_id"`
	ScheduleName string `json:"schedule_name"`
	UserID       string `json:"user_id"`
	IsActive     bool   `json:"is_active"`
}

type DB struct {
	db *sql.DB

	epID *sql.Stmt

	isOnCall *sql.Stmt

	onCallEPDirect *sql.Stmt
	onCallEPRot    *sql.Stmt

	onCallRemoveOverrides            *sql.Stmt
	onCallAddOverrideAssignments     *sql.Stmt
	onCallReplaceOverrideAssignments *sql.Stmt
	onCallDirectAssignments          *sql.Stmt

	rules rule.Store
	sched schedule.Store
}

func NewDB(ctx context.Context, db *sql.DB, rules rule.Store, sched schedule.Store) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &DB{
		db: db,

		rules: rules,
		sched: sched,

		isOnCall: p.P(`
			select 1
			from services svc
			join escalation_policy_steps step on step.escalation_policy_id = svc.escalation_policy_id
			join escalation_policy_actions act on act.escalation_policy_step_id = step.id
			left join schedule_on_call_users sUser on
				sUser.schedule_id = act.schedule_id and
				sUser.user_id = $1 and
				sUser.end_time isnull
			left join rotation_state rState on rState.rotation_id = act.rotation_id
			left join rotation_participants part on
				part.id = rState.rotation_participant_id and
				part.user_id = $1
			where coalesce(act.user_id, sUser.user_id, part.user_id) = $1
			limit 1
		`),

		onCallEPDirect: p.P(`
			select distinct
				svc.id,
				svc.name,
				ep.id,
				ep.name,
				step.step_number,
				null,
				null,
				null,
				null,
				true
			from services svc
			join escalation_policies ep on ep.id = svc.escalation_policy_id
			join escalation_policy_steps step on step.escalation_policy_id = svc.escalation_policy_id
			join escalation_policy_actions act on
				act.escalation_policy_step_id = step.id and
				act.user_id = $1
		`),

		onCallEPRot: p.P(`
			select distinct
				svc.id,
				svc.name,
				ep.id,
				ep.name,
				step.step_number,
				null,
				null,
				act.rotation_id,
				rot.name,
				part.id = state.rotation_participant_id
			from services svc
			join escalation_policies ep on ep.id = svc.escalation_policy_id
			join escalation_policy_steps step on step.escalation_policy_id = svc.escalation_policy_id
			join escalation_policy_actions act on
				act.escalation_policy_step_id = step.id and
				act.rotation_id notnull
			join rotations rot on rot.id = act.rotation_id
			join rotation_participants part on
				part.user_id = $1 and
				part.rotation_id = act.rotation_id
			join rotation_state state on state.rotation_id = act.rotation_id
		`),

		onCallRemoveOverrides: p.P(`
			select
				tgt_schedule_id
			from user_overrides
			where
				now() between start_time and end_time and
				remove_user_id = $1
		`),

		onCallAddOverrideAssignments: p.P(`
			select
				svc.id,
				svc.name,
				ep.id,
				ep.name,
				step.step_number,
				sched.id,
				sched.name,
				null,
				null,
				o.start_time <= now()
			from user_overrides o
			join schedules sched on sched.id = o.tgt_schedule_id
			join escalation_policy_actions act on act.schedule_id = o.tgt_schedule_id
			join escalation_policy_steps step on step.id = act.escalation_policy_step_id
			join escalation_policies ep on ep.id = step.escalation_policy_id
			join services svc on svc.escalation_policy_id = step.escalation_policy_id
			where
				o.end_time > now() and
				o.add_user_id = $1 and
				o.remove_user_id isnull
		`),
		onCallReplaceOverrideAssignments: p.P(`
			select distinct
				svc.id,
				svc.name,
				ep.id,
				ep.name,
				step.step_number,
				sched.id,
				sched.name,
				null,
				null,
				onCall.user_id notnull
			from user_overrides o
			join escalation_policy_actions act on act.schedule_id = o.tgt_schedule_id
			join escalation_policy_steps step on step.id = act.escalation_policy_step_id
			join escalation_policies ep on ep.id = step.escalation_policy_id
			join services svc on svc.escalation_policy_id = step.escalation_policy_id
			join schedules sched on sched.id = o.tgt_schedule_id
			join schedule_rules rule on
				rule.schedule_id = o.tgt_schedule_id and
				(rule.tgt_user_id isnull or rule.tgt_user_id = o.remove_user_id)
			left join schedule_on_call_users onCall on
				onCall.user_id = $1 and
				onCall.schedule_id = act.schedule_id and
				onCall.end_time isnull
			left join rotation_participants part on
				part.rotation_id = rule.tgt_rotation_id and
				part.user_id = o.remove_user_id
			where
				o.end_time > now() and
				o.add_user_id = $1 and
				o.remove_user_id notnull and
				(rule.tgt_user_id notnull or part notnull)
		`),

		onCallDirectAssignments: p.P(`
			select distinct
				svc.id,
				svc.name,
				ep.id,
				ep.name,
				step.step_number,
				sched.id,
				sched.name,
				null,
				null,
				onCall.user_id notnull
			from services svc
			join escalation_policies ep on ep.id = svc.escalation_policy_id
			join escalation_policy_steps step on step.escalation_policy_id = svc.escalation_policy_id
			join escalation_policy_actions act on
				act.escalation_policy_step_id = step.id and
				act.schedule_id notnull
			join schedule_rules rule on
				rule.schedule_id = act.schedule_id and
				(rule.tgt_user_id isnull or rule.tgt_user_id = $1)
			join schedules sched on sched.id = act.schedule_id
			left join schedule_on_call_users onCall on
				onCall.user_id = $1 and
				onCall.schedule_id = act.schedule_id and
				onCall.end_time isnull
			left join rotation_participants part on
				part.user_id = $1 and
				part.rotation_id = rule.tgt_rotation_id
			where
				rule.tgt_user_id notnull or
				part notnull
		`),

		epID: p.P(`
			select escalation_policy_id
			from alerts a
			join services s
				on s.id = a.service_id
			where a.id = $1
		`),
	}, p.Err
}

func (db *DB) AlertEPID(ctx context.Context, alertID int) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return "", err
	}

	row := db.epID.QueryRowContext(ctx, alertID)
	var epID string
	return epID, errors.Wrap(row.Scan(&epID), "query alert EP ID")
}

func (db *DB) IsUserOnCall(ctx context.Context, userID string) (bool, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return false, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return false, err
	}
	var result int
	err = db.isOnCall.QueryRowContext(ctx, userID).Scan(&result)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

type assignment struct {
	ServiceID   string
	ServiceName string
	EPID        string
	EPName      string
	StepNumber  int
	SchedID     sql.NullString
	SchedName   sql.NullString
	RotID       sql.NullString
	RotName     sql.NullString
	Active      bool
}

func (a assignment) ID() assignmentID {
	return assignmentID{
		ServiceID:  a.ServiceID,
		EPID:       a.EPID,
		StepNumber: a.StepNumber,
		SchedID:    a.SchedID.String,
		RotID:      a.RotID.String,
	}
}

type assignmentID struct {
	ServiceID  string
	EPID       string
	StepNumber int
	SchedID    string
	RotID      string
}

func (db *DB) OnCallByUser(ctx context.Context, userID string) ([]OnCallAssignment, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	fetchAll := func(s *sql.Stmt) ([]assignment, error) {
		var result []assignment
		rows, err := tx.Stmt(s).QueryContext(ctx, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var f assignment
		for rows.Next() {
			err = rows.Scan(&f.ServiceID, &f.ServiceName, &f.EPID, &f.EPName, &f.StepNumber, &f.SchedID, &f.SchedName, &f.RotID, &f.RotName, &f.Active)
			if err != nil {
				return nil, err
			}
			result = append(result, f)
		}
		return result, nil
	}

	direct, err := fetchAll(db.onCallDirectAssignments)
	if err != nil {
		return nil, errors.Wrap(err, "fetch direct")
	}
	replace, err := fetchAll(db.onCallReplaceOverrideAssignments)
	if err != nil {
		return nil, errors.Wrap(err, "fetch replace")
	}
	add, err := fetchAll(db.onCallAddOverrideAssignments)
	if err != nil {
		return nil, errors.Wrap(err, "fetch add")
	}
	userTgt, err := fetchAll(db.onCallEPDirect)
	if err != nil {
		return nil, errors.Wrap(err, "fetch ep direct")
	}
	rotTgt, err := fetchAll(db.onCallEPRot)
	if err != nil {
		return nil, errors.Wrap(err, "fetch ep rotations")
	}
	remove := make(map[string]bool)
	rows, err := tx.Stmt(db.onCallRemoveOverrides).QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var id string
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		remove[id] = true
	}
	// done with DB stuff
	rows.Close()
	tx.Rollback()

	m := make(map[assignmentID]assignment)

	all := append(direct, replace...)
	all = append(all, add...)
	all = append(all, userTgt...)
	all = append(all, rotTgt...)

	for _, a := range all {
		id := a.ID()
		e := m[id]
		a.Active = (e.Active || a.Active) && !remove[a.SchedID.String]
		m[id] = a
	}

	result := make([]OnCallAssignment, 0, len(m))

	for _, a := range m {
		result = append(result, OnCallAssignment{
			ServiceID:    a.ServiceID,
			ServiceName:  a.ServiceName,
			EPID:         a.EPID,
			EPName:       a.EPName,
			Level:        a.StepNumber,
			RotationID:   a.RotID.String,
			RotationName: a.RotName.String,
			ScheduleID:   a.SchedID.String,
			ScheduleName: a.SchedName.String,
			UserID:       userID,
			IsActive:     a.Active,
		})
	}

	return result, nil
}
