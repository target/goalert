package oncall

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/override"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/schedule/rule"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"
)

// ScheduleOnCallUser represents a currently on-call user for a schedule.
type ScheduleOnCallUser struct {
	ID   string
	Name string
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

// Store allows retrieving and calculating on-call information.
type Store struct {
	db *sql.DB

	onCallUsersSvc      *sql.Stmt
	onCallUsersSchedule *sql.Stmt
	schedOverrides      *sql.Stmt

	schedOnCall *sql.Stmt
	schedTZ     *sql.Stmt
	schedRot    *sql.Stmt
	rotParts    *sql.Stmt

	ruleStore  *rule.Store
	schedStore *schedule.Store
}

// NewStore will create a new DB, preparing required statements using the provided context.
func NewStore(ctx context.Context, db *sql.DB, ruleStore *rule.Store, schedStore *schedule.Store) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db:         db,
		ruleStore:  ruleStore,
		schedStore: schedStore,

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
		onCallUsersSchedule: p.P(`
			SELECT s.user_id, u.name
			FROM schedule_on_call_users s
			JOIN users u ON u.id = s.user_id
			WHERE s.schedule_id = $1 AND s.end_time IS NULL
		`),
		schedOnCall: p.P(`
			select
				user_id,
				start_time,
				end_time
			from schedule_on_call_users
			where
				schedule_id = $1 and
				tstzrange($2, $3) && tstzrange(start_time, end_time) and
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
func (s *Store) OnCallUsersByService(ctx context.Context, serviceID string) ([]ServiceOnCallUser, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ServiceID", serviceID)
	if err != nil {
		return nil, err
	}
	rows, err := s.onCallUsersSvc.QueryContext(ctx, serviceID)
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

func (s *Store) OnCallUsersBySchedule(ctx context.Context, scheduleID string) ([]ScheduleOnCallUser, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	rows, err := s.onCallUsersSchedule.QueryContext(ctx, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("fetch on-call users for schedule '%s': %w", scheduleID, err)
	}
	defer rows.Close()

	var result []ScheduleOnCallUser
	for rows.Next() {
		var u ScheduleOnCallUser
		err = rows.Scan(&u.ID, &u.Name)
		if err != nil {
			return nil, fmt.Errorf("scan on-call user entry #%d for schedule '%s': %w", len(result), scheduleID, err)
		}

		result = append(result, u)
	}

	return result, nil
}

func filterShifts(s []Shift, userID string) []Shift {
	var shifts []Shift
	for _, shift := range s {
		if shift.UserID == userID {
			shifts = append(shifts, shift)
		}
	}
	return shifts
}

func (s *Store) ShiftsByUser(ctx context.Context, scheduleID string, start, end time.Time, userID string) ([]Shift, error) {
	shifts, err := s.HistoryBySchedule(ctx, scheduleID, start, end)
	if err != nil {
		return nil, err
	}

	return filterShifts(shifts, userID), nil
}

// HistoryBySchedule will return the list of shifts that overlap the start and end time for the given schedule.
func (s *Store) HistoryBySchedule(ctx context.Context, scheduleID string, start, end time.Time) ([]Shift, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		ReadOnly:  true,
		Isolation: sql.LevelRepeatableRead,
	})
	if err != nil {
		return nil, errors.Wrap(err, "begin transaction")
	}
	defer sqlutil.Rollback(ctx, "oncall: fetch schedule history", tx)

	var schedTZ string
	var now time.Time
	err = tx.StmtContext(ctx, s.schedTZ).QueryRowContext(ctx, scheduleID).Scan(&schedTZ, &now)
	if err != nil {
		return nil, errors.Wrap(err, "lookup schedule time zone")
	}

	rows, err := tx.StmtContext(ctx, s.schedRot).QueryContext(ctx, scheduleID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup schedule rotations")
	}
	defer rows.Close()
	rots := make(map[string]*ResolvedRotation)
	var rotIDs []string
	for rows.Next() {
		var rot ResolvedRotation
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

	rows, err = tx.StmtContext(ctx, s.rotParts).QueryContext(ctx, sqlutil.UUIDArray(rotIDs))
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

	rawRules, err := s.ruleStore.FindAllTx(ctx, tx, scheduleID)
	if err != nil {
		return nil, errors.Wrap(err, "lookup schedule rules")
	}

	var rules []ResolvedRule
	for _, r := range rawRules {
		if r.Target.TargetType() == assignment.TargetTypeRotation {
			rules = append(rules, ResolvedRule{
				Rule:     r,
				Rotation: rots[r.Target.TargetID()],
			})
		} else {
			rules = append(rules, ResolvedRule{Rule: r})
		}
	}

	rows, err = tx.StmtContext(ctx, s.schedOnCall).QueryContext(ctx, scheduleID, start, end)
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

	rows, err = tx.StmtContext(ctx, s.schedOverrides).QueryContext(ctx, scheduleID, start, end)
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
	id, err := uuid.Parse(scheduleID)
	if err != nil {
		return nil, errors.Wrap(err, "parse schedule ID")
	}
	tempScheds, err := s.schedStore.TemporarySchedules(ctx, tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "lookup temporary schedules")
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
	st := state{
		rules:      rules,
		overrides:  overrides,
		history:    userHistory,
		now:        now,
		loc:        tz,
		tempScheds: tempScheds,
	}

	return st.CalculateShifts(start, end), nil
}
