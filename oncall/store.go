package oncall

import (
	"context"
	"database/sql"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
	"time"

	"github.com/pkg/errors"
)

// Store allows retrieving and calculating on-call information.
type Store interface {
	OnCallUsersByService(ctx context.Context, serviceID string) ([]ServiceOnCallUser, error)

	// HistoryBySchedule(ctx context.Context, stepID string, start, end time.Time) ([]Shift, error)
	HistoryBySchedule(ctx context.Context, scheduleID string, start, end time.Time) ([]Shift, error)
}

// ServiceOnCallUser represents a currently on-call user for a service.
type ServiceOnCallUser struct {
	StepNumber int    `json:"step_number"`
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
}

// A Shift represents a duration a user is on-call.
// If truncated is true, then the End time does not represent
// the time the user stopped being on call, instead it indicates
// they were still on-call at that time.
type Shift struct {
	UserID    string    `json:"user_id"`
	Start     time.Time `json:"start_time"`
	End       time.Time `json:"end_time"`
	Truncated bool      `json:"truncated"`
}

// DB implements the Store interface from Postgres.
type DB struct {
	db *sql.DB

	onCallUsersSvc *sql.Stmt
	schedOverrides *sql.Stmt

	schedOnCall *sql.Stmt
	schedTZ     *sql.Stmt
	schedRot    *sql.Stmt
	rotParts    *sql.Stmt

	ruleStore rule.Store
}

// NewDB will create a new DB, preparing required statements using the provided context.
func NewDB(ctx context.Context, db *sql.DB, ruleStore rule.Store) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		db:        db,
		ruleStore: ruleStore,

		schedOverrides: p.P(`
			select
				start_time,
				end_time,
				add_user_id,
				remove_user_id
			from user_overrides
			where
				tgt_schedule_id = $1 and
				end_time > now() and
				($2, $3) OVERLAPS(start_time, end_time)
		`),

		onCallUsersSvc: p.P(`
			select step.step_number, oc.user_id, u.name as user_name
			from services svc
			join escalation_policy_steps step on step.escalation_policy_id = svc.escalation_policy_id
			join ep_step_on_call_users oc on oc.ep_step_id = step.id and oc.end_time isnull
			join users u on oc.user_id = u.id
			where svc.id = $1
			order by step.step_number, oc.start_time
		`),

		schedOnCall: p.P(`
			select
				user_id,
				start_time,
				end_time
			from schedule_on_call_users
			where
				schedule_id = $1 and
				($2, $3) OVERLAPS (start_time, coalesce(end_time, 'infinity')) and
				(end_time isnull or (end_time - start_time) > '1 minute'::interval)
		`),
		schedTZ: p.P(`select time_zone, now() from schedules where id = $1`),
		schedRot: p.P(`
			select distinct
				rot.id,
				rot.type,
				rot.start_time,
				rot.shift_length,
				rot.time_zone,
				state.position,
				state.shift_start
			from schedule_rules rule
			join rotations rot on rot.id = rule.tgt_rotation_id
			join rotation_state state on state.rotation_id = rule.tgt_rotation_id
			where rule.schedule_id = $1 and rule.tgt_rotation_id notnull
		`),
		rotParts: p.P(`
			select
				rotation_id,
				user_id
			from rotation_participants
			where rotation_id = any($1)
			order by
				rotation_id,
				position
		`),
	}, p.Err
}

// OnCallUsersByService will return the current set of users who are on-call for the given service.
func (db *DB) OnCallUsersByService(ctx context.Context, serviceID string) ([]ServiceOnCallUser, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}
	rows, err := db.onCallUsersSvc.QueryContext(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var onCall []ServiceOnCallUser
	for rows.Next() {
		var u ServiceOnCallUser
		err = rows.Scan(&u.StepNumber, &u.UserID, &u.UserName)
		if err != nil {
			return nil, err
		}
		onCall = append(onCall, u)
	}
	return onCall, nil
}

// HistoryBySchedule will return the list of shifts that overlap the start and end time for the given schedule.
func (db *DB) HistoryBySchedule(ctx context.Context, scheduleID string, start, end time.Time) ([]Shift, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}

	tx, err := db.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly:  true,
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return nil, errors.Wrap(err, "begin transaction")
	}
	defer tx.Rollback()

	var schedTZ string
	var now time.Time
	err = tx.StmtContext(ctx, db.schedTZ).QueryRowContext(ctx, scheduleID).Scan(&schedTZ, &now)
	if err != nil {
		return nil, errors.Wrap(err, "lookup schedule time zone")
	}

	rows, err := tx.StmtContext(ctx, db.schedRot).QueryContext(ctx, scheduleID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup schedule rotations")
	}
	defer rows.Close()
	rots := make(map[string]*resolvedRotation)
	var rotIDs []string
	for rows.Next() {
		var rot resolvedRotation
		var rotTZ string
		err = rows.Scan(&rot.ID, &rot.Type, &rot.Start, &rot.ShiftLength, &rotTZ, &rot.CurrentIndex, &rot.CurrentStart)
		if err != nil {
			return nil, errors.Wrap(err, "scan rotation info")
		}
		loc, err := util.LoadLocation(rotTZ)
		if err != nil {
			return nil, errors.Wrap(err, "load time zone info")
		}
		rot.Start = rot.Start.In(loc)
		rots[rot.ID] = &rot
		rotIDs = append(rotIDs, rot.ID)
	}

	rows, err = tx.StmtContext(ctx, db.rotParts).QueryContext(ctx, sqlutil.UUIDArray(rotIDs))
	if err != nil {
		return nil, errors.Wrap(err, "lookup rotation participants")
	}
	defer rows.Close()
	for rows.Next() {
		var rotID, userID string
		err = rows.Scan(&rotID, &userID)
		if err != nil {
			return nil, errors.Wrap(err, "scan rotation participant info")
		}
		rots[rotID].Users = append(rots[rotID].Users, userID)
	}

	rawRules, err := db.ruleStore.FindAllTx(ctx, tx, scheduleID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup schedule rules")
	}

	var rules []resolvedRule
	for _, r := range rawRules {
		if r.Target.TargetType() == assignment.TargetTypeRotation {
			rules = append(rules, resolvedRule{
				Rule:     r,
				Rotation: rots[r.Target.TargetID()],
			})
		} else {
			rules = append(rules, resolvedRule{Rule: r})
		}
	}

	rows, err = tx.StmtContext(ctx, db.schedOnCall).QueryContext(ctx, scheduleID, start, end)
	if err != nil {
		return nil, errors.Wrap(err, "lookup on-call history")
	}
	defer rows.Close()
	var userHistory []Shift
	for rows.Next() {
		var s Shift
		var end sqlutil.NullTime
		err = rows.Scan(&s.UserID, &s.Start, &end)
		if err != nil {
			return nil, errors.Wrap(err, "scan on-call history info")
		}
		s.End = end.Time
		userHistory = append(userHistory, s)
	}

	rows, err = tx.StmtContext(ctx, db.schedOverrides).QueryContext(ctx, scheduleID, start, end)
	if err != nil {
		return nil, errors.Wrap(err, "lookup overrides")
	}
	defer rows.Close()
	var overrides []override.UserOverride
	for rows.Next() {
		var add, rem sql.NullString
		var ov override.UserOverride
		err = rows.Scan(&ov.Start, &ov.End, &add, &rem)
		if err != nil {
			return nil, errors.Wrap(err, "scan override info")
		}
		ov.AddUserID = add.String
		ov.RemoveUserID = rem.String
		overrides = append(overrides, ov)
	}

	err = tx.Commit()
	if err != nil {
		// Can't use the data we read (e.g. serialization error)
		return nil, errors.Wrap(err, "commit tx")
	}
	tz, err := util.LoadLocation(schedTZ)
	if err != nil {
		return nil, errors.Wrap(err, "load time zone info")
	}
	s := state{
		rules:     rules,
		overrides: overrides,
		history:   userHistory,
		now:       now,
		loc:       tz,
	}

	return s.CalculateShifts(start, end), nil
}
