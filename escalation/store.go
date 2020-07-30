package escalation

import (
	"context"
	"database/sql"

	"github.com/target/goalert/util/sqlutil"

	alertlog "github.com/target/goalert/alert/log"
	"github.com/target/goalert/assignment"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/validation/validate"

	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

type Store interface {
	PolicyStore
	StepStore
	ActiveStepReader
}

type PolicyStore interface {
	FindOnePolicy(context.Context, string) (*Policy, error)
	FindOnePolicyTx(context.Context, *sql.Tx, string) (*Policy, error)
	FindOnePolicyForUpdateTx(context.Context, *sql.Tx, string) (*Policy, error)
	FindAllPolicies(context.Context) ([]Policy, error)
	CreatePolicy(context.Context, *Policy) (*Policy, error)
	CreatePolicyTx(context.Context, *sql.Tx, *Policy) (*Policy, error)
	UpdatePolicy(context.Context, *Policy) error
	UpdatePolicyTx(context.Context, *sql.Tx, *Policy) error
	DeletePolicy(ctx context.Context, id string) error
	DeletePolicyTx(ctx context.Context, tx *sql.Tx, id string) error
	FindAllStepTargets(ctx context.Context, stepID string) ([]assignment.Target, error)
	FindAllStepTargetsTx(ctx context.Context, tx *sql.Tx, stepID string) ([]assignment.Target, error)
	AddStepTarget(ctx context.Context, stepID string, tgt assignment.Target) error
	AddStepTargetTx(ctx context.Context, tx *sql.Tx, stepID string, tgt assignment.Target) error
	DeleteStepTarget(ctx context.Context, stepID string, tgt assignment.Target) error
	DeleteStepTargetTx(ctx context.Context, tx *sql.Tx, stepID string, tgt assignment.Target) error
	FindAllPoliciesBySchedule(ctx context.Context, scheduleID string) ([]Policy, error)
	FindManyPolicies(ctx context.Context, ids []string) ([]Policy, error)
	DeleteManyPoliciesTx(ctx context.Context, tx *sql.Tx, ids []string) error
	FindAllNotices(ctx context.Context, policyID string) ([]Notice, error)

	Search(context.Context, *SearchOptions) ([]Policy, error)
}

type StepStore interface {
	// CreateStep is replaced by CreateStepTx.
	CreateStep(context.Context, *Step) (*Step, error)

	// CreateStepTx will create an escalation policy step within the given transaction.
	// Note: Targets are not assigned during creation.
	CreateStepTx(context.Context, *sql.Tx, *Step) (*Step, error)
	UpdateStepNumberTx(context.Context, *sql.Tx, string, int) error

	// Update step allows updating a steps delay
	// Note: it does not update the Targets.
	UpdateStep(context.Context, *Step) error
	UpdateStepDelayTx(context.Context, *sql.Tx, string, int) error
	DeleteStep(context.Context, string) (string, error)
	DeleteStepTx(context.Context, *sql.Tx, string) (string, error)
	MoveStep(context.Context, string, int) error
}

type ActiveStepReader interface {
	ActiveStep(ctx context.Context, alertID int, policyID string) (*ActiveStep, error)

	// FindOneStep will return a single escalation policy step.
	// Note: it does not currently fetch the Targets.
	FindOneStep(context.Context, string) (*Step, error)
	FindOneStepTx(context.Context, *sql.Tx, string) (*Step, error)
	FindOneStepForUpdateTx(context.Context, *sql.Tx, string) (*Step, error)

	// FindAllSteps will return escalation policy steps for the given policy ID.
	// Note: it does not currently fetch the Targets.
	FindAllSteps(context.Context, string) ([]Step, error)
	FindAllStepsTx(context.Context, *sql.Tx, string) ([]Step, error)
	FindAllOnCallStepsForUserTx(ctx context.Context, tx *sql.Tx, userID string) ([]Step, error)
}

type Manager interface {
	ActiveStepReader

	FindOnePolicy(context.Context, string) (*Policy, error)
}

var _ Manager = &DB{}
var _ Store = &DB{}

type Config struct {
	NCStore         notificationchannel.Store
	LogStore        alertlog.Store
	SlackLookupFunc func(ctx context.Context, channelID string) (*slack.Channel, error)
}

type DB struct {
	db *sql.DB

	log     alertlog.Store
	ncStore notificationchannel.Store
	slackFn func(ctx context.Context, channelID string) (*slack.Channel, error)

	findSlackChan *sql.Stmt

	findOnePolicy             *sql.Stmt
	findOnePolicyForUpdate    *sql.Stmt
	findManyPolicies          *sql.Stmt
	findAllPolicies           *sql.Stmt
	findAllPoliciesBySchedule *sql.Stmt
	createPolicy              *sql.Stmt
	updatePolicy              *sql.Stmt
	deletePolicy              *sql.Stmt

	findOneStep          *sql.Stmt
	findOneStepForUpdate *sql.Stmt
	findAllSteps         *sql.Stmt
	findAllOnCallSteps   *sql.Stmt
	createStep           *sql.Stmt
	updateStepDelay      *sql.Stmt
	updateStepNumber     *sql.Stmt
	deleteStep           *sql.Stmt
	moveStep             *sql.Stmt

	activeStep *sql.Stmt

	addStepTarget      *sql.Stmt
	deleteStepTarget   *sql.Stmt
	findAllStepTargets *sql.Stmt
}

func NewDB(ctx context.Context, db *sql.DB, cfg Config) (*DB, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &DB{
		db:      db,
		log:     cfg.LogStore,
		slackFn: cfg.SlackLookupFunc,
		ncStore: cfg.NCStore,

		findSlackChan: p.P(`
			SELECT chan.id
			FROM notification_channels chan
			JOIN escalation_policy_actions act ON
				act.escalation_policy_step_id = $1 AND
				act.channel_id = chan.id
			WHERE chan.value = $2
		`),

		findOnePolicy:          p.P(`SELECT id, name, description, repeat FROM escalation_policies WHERE id = $1`),
		findOnePolicyForUpdate: p.P(`SELECT id, name, description, repeat FROM escalation_policies WHERE id = $1 FOR UPDATE`),
		findManyPolicies:       p.P(`SELECT id, name, description, repeat FROM escalation_policies WHERE id = any($1)`),

		findAllPolicies: p.P(`SELECT id, name, description, repeat FROM escalation_policies`),
		findAllPoliciesBySchedule: p.P(`
			SELECT DISTINCT
				step.escalation_policy_id,
				pol.name,
				pol.description,
				pol.repeat
			FROM
				escalation_policy_actions as act
			JOIN
				escalation_policy_steps as step on step.id = act.escalation_policy_step_id
			JOIN
				escalation_policies as pol on pol.id = step.escalation_policy_id
			WHERE
				act.schedule_id = $1
		`),
		createPolicy: p.P(`INSERT INTO escalation_policies (id, name, description, repeat) VALUES ($1, $2, $3, $4)`),
		updatePolicy: p.P(`UPDATE escalation_policies SET name = $2, description = $3, repeat = $4 WHERE id = $1`),
		deletePolicy: p.P(`DELETE FROM escalation_policies WHERE id = any($1)`),

		addStepTarget: p.P(`
			INSERT INTO escalation_policy_actions (id, escalation_policy_step_id, user_id, schedule_id, rotation_id, channel_id)
			VALUES ($1, $2, $3, $4, $5, $6)
		`),
		deleteStepTarget: p.P(`
			DELETE FROM escalation_policy_actions
			WHERE
				escalation_policy_step_id = $1 AND
				(
					user_id = $2 OR
					schedule_id = $3 OR
					rotation_id = $4 OR
					channel_id = $5
				)
		`),
		findAllStepTargets: p.P(`
			SELECT
				user_id,
				schedule_id,
				rotation_id,
				channel_id,
				chan.type,
				chan.value,
				COALESCE(users.name, rot.name, sched.name, chan.name)
			FROM
				escalation_policy_actions act
			LEFT JOIN users
				on act.user_id = users.id
			LEFT JOIN rotations rot
				on act.rotation_id = rot.id
			LEFT JOIN schedules sched
				on act.schedule_id = sched.id
			LEFT JOIN notification_channels chan
				on act.channel_id = chan.id
			WHERE
				escalation_policy_step_id = $1
		`),

		findOneStep:          p.P(`SELECT id, escalation_policy_id, delay, step_number FROM escalation_policy_steps WHERE id = $1`),
		findOneStepForUpdate: p.P(`SELECT id, escalation_policy_id, delay, step_number FROM escalation_policy_steps WHERE id = $1 FOR UPDATE`),
		findAllSteps:         p.P(`SELECT id, escalation_policy_id, delay, step_number FROM escalation_policy_steps WHERE escalation_policy_id = $1 ORDER BY step_number`),
		findAllOnCallSteps: p.P(`
			SELECT step.id, step.escalation_policy_id, step.delay, step.step_number
			FROM ep_step_on_call_users oc
			JOIN escalation_policy_steps step ON step.id = oc.ep_step_id
			WHERE oc.user_id = $1 AND oc.end_time isnull
			ORDER BY step.escalation_policy_id, step.step_number
		`),

		createStep: p.P(`
			INSERT INTO escalation_policy_steps
				(id, escalation_policy_id, delay, step_number)
			VALUES ($1, $2, $3, DEFAULT)
			RETURNING step_number
		`),
		updateStepDelay:  p.P(`UPDATE escalation_policy_steps SET delay = $2 WHERE id = $1`),
		updateStepNumber: p.P(`UPDATE escalation_policy_steps SET step_number = $2 WHERE id = $1`),
		deleteStep:       p.P(`DELETE FROM escalation_policy_steps WHERE id = $1 RETURNING escalation_policy_id`),
		moveStep: p.P(`
			WITH calc AS (
				SELECT
					escalation_policy_id esc_id,
					step_number old_pos,
					LEAST(step_number, $2) min,
					GREATEST(step_number, $2) max,
					($2 - step_number) diff,
					CASE
						WHEN step_number < $2 THEN abs($2-step_number)
						WHEN step_number > $2 THEN 1
						ELSE 0
					END shift
				FROM escalation_policy_steps
				WHERE id = $1
				FOR UPDATE
			)
			UPDATE escalation_policy_steps
			SET step_number =  ((step_number - calc.min) + calc.shift) % (abs(calc.diff) + 1) + calc.min
			FROM calc
			WHERE
				escalation_policy_id = calc.esc_id AND
				step_number >= calc.min AND
				step_number <= calc.max
			RETURNING escalation_policy_id
		`),

		activeStep: p.P(`
			SELECT
				escalation_policy_step_id,
				last_escalation,
				loop_count,
				force_escalation,
				escalation_policy_step_number
			FROM escalation_policy_state
			WHERE alert_id = $1 AND escalation_policy_id = $2
		`),
	}, p.Err
}

func (db *DB) logChange(ctx context.Context, tx *sql.Tx, policyID string) {
	err := db.log.LogEPTx(ctx, tx, policyID, alertlog.TypePolicyUpdated, nil)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "append alertlog (escalation policy update)"))
	}
}

func validStepTarget(tgt assignment.Target) error {
	return validate.Many(
		validate.UUID("TargetID", tgt.TargetID()),
		validate.OneOf("TargetType", tgt.TargetType(),
			assignment.TargetTypeUser,
			assignment.TargetTypeSchedule,
			assignment.TargetTypeRotation,
			assignment.TargetTypeNotificationChannel,
		),
	)
}

func tgtFields(id string, tgt assignment.Target, insert bool) []interface{} {
	var usr, sched, rot, ch sql.NullString
	switch tgt.TargetType() {
	case assignment.TargetTypeUser:
		usr.Valid = true
		usr.String = tgt.TargetID()
	case assignment.TargetTypeSchedule:
		sched.Valid = true
		sched.String = tgt.TargetID()
	case assignment.TargetTypeRotation:
		rot.Valid = true
		rot.String = tgt.TargetID()
	case assignment.TargetTypeNotificationChannel:
		ch.Valid = true
		ch.String = tgt.TargetID()
	}
	if insert {
		return []interface{}{
			uuid.NewV4().String(),
			id,
			usr,
			sched,
			rot,
			ch,
		}
	}
	return []interface{}{
		id,
		usr,
		sched,
		rot,
		ch,
	}
}

func (db *DB) FindManyPolicies(ctx context.Context, ids []string) ([]Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.ManyUUID("EscalationPolicyID", ids, 200)
	if err != nil {
		return nil, err
	}

	rows, err := db.findManyPolicies.QueryContext(ctx, sqlutil.UUIDArray(ids))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Policy
	var p Policy
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}

func (db *DB) _updateStepTarget(ctx context.Context, stepID string, tgt assignment.Target, stmt *sql.Stmt, insert bool) error {
	err := validate.Many(
		validate.UUID("StepID", stepID),
		validStepTarget(tgt),
	)
	if err != nil {
		return err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, tgtFields(stepID, tgt, insert)...)
	if err == sql.ErrNoRows {
		err = nil
	}
	return err
}

func (db *DB) newSlackChannel(ctx context.Context, tx *sql.Tx, slackChanID string) (assignment.Target, error) {
	ch, err := db.slackFn(ctx, slackChanID)
	if err != nil {
		return nil, err
	}

	notifCh, err := db.ncStore.CreateTx(ctx, tx, &notificationchannel.Channel{
		Type:  notificationchannel.TypeSlack,
		Name:  ch.Name,
		Value: ch.ID,
	})
	if err != nil {
		return nil, err
	}

	return assignment.NotificationChannelTarget(notifCh.ID), nil
}
func (db *DB) lookupSlackChannel(ctx context.Context, tx *sql.Tx, stepID, slackChanID string) (assignment.Target, error) {
	var notifChanID string
	err := tx.StmtContext(ctx, db.findSlackChan).QueryRowContext(ctx, stepID, slackChanID).Scan(&notifChanID)
	if err != nil {
		return nil, err
	}

	return assignment.NotificationChannelTarget(notifChanID), nil
}

func (db *DB) AddStepTarget(ctx context.Context, stepID string, tgt assignment.Target) error {
	return db._updateStepTarget(ctx, stepID, tgt, db.addStepTarget, true)
}

func (db *DB) AddStepTargetTx(ctx context.Context, tx *sql.Tx, stepID string, tgt assignment.Target) error {
	if tgt.TargetType() == assignment.TargetTypeSlackChannel {
		var err error
		tgt, err = db.newSlackChannel(ctx, tx, tgt.TargetID())
		if err != nil {
			return err
		}
	}
	return db._updateStepTarget(ctx, stepID, tgt, tx.StmtContext(ctx, db.addStepTarget), true)
}

func (db *DB) DeleteStepTarget(ctx context.Context, stepID string, tgt assignment.Target) error {
	return db._updateStepTarget(ctx, stepID, tgt, db.deleteStepTarget, false)
}

func (db *DB) DeleteStepTargetTx(ctx context.Context, tx *sql.Tx, stepID string, tgt assignment.Target) error {
	if tgt.TargetType() == assignment.TargetTypeSlackChannel {
		var err error
		tgt, err = db.lookupSlackChannel(ctx, tx, stepID, tgt.TargetID())
		if err != nil {
			return err
		}
	}
	return db._updateStepTarget(ctx, stepID, tgt, tx.StmtContext(ctx, db.deleteStepTarget), false)
}

func (db *DB) FindAllStepTargets(ctx context.Context, stepID string) ([]assignment.Target, error) {
	return db.FindAllStepTargetsTx(ctx, nil, stepID)
}

func (db *DB) FindAllStepTargetsTx(ctx context.Context, tx *sql.Tx, stepID string) ([]assignment.Target, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("StepID", stepID)
	if err != nil {
		return nil, err
	}

	stmt := db.findAllStepTargets
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	rows, err := stmt.QueryContext(ctx, stepID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tgts []assignment.Target
	for rows.Next() {
		var usr, sched, rot, ch, chValue sql.NullString
		var chType *notificationchannel.Type
		var tgt assignment.RawTarget
		err = rows.Scan(&usr, &sched, &rot, &ch, &chType, &chValue, &tgt.Name)
		if err != nil {
			return nil, err
		}

		switch {
		case usr.Valid:
			tgt.ID = usr.String
			tgt.Type = assignment.TargetTypeUser
		case sched.Valid:
			tgt.ID = sched.String
			tgt.Type = assignment.TargetTypeSchedule
		case rot.Valid:
			tgt.ID = rot.String
			tgt.Type = assignment.TargetTypeRotation
		case ch.Valid:
			switch *chType {
			case notificationchannel.TypeSlack:
				tgt.ID = chValue.String
				tgt.Type = assignment.TargetTypeSlackChannel
			default:
				tgt.ID = ch.String
				tgt.Type = assignment.TargetTypeNotificationChannel
			}
		default:
			continue
		}
		tgts = append(tgts, tgt)
	}

	return tgts, nil
}

func (db *DB) ActiveStep(ctx context.Context, alertID int, policyID string) (*ActiveStep, error) {
	err := validate.UUID("EscalationPolicyID", policyID)
	if err != nil {
		return nil, err
	}
	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	row := db.activeStep.QueryRowContext(ctx, alertID, policyID)
	var step ActiveStep
	var stepID sql.NullString
	err = row.Scan(&stepID, &step.LastEscalation, &step.LoopCount, &step.ForceEscalation, &step.StepNumber)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	step.StepID = stepID.String
	step.PolicyID = policyID
	step.AlertID = alertID
	return &step, nil
}

func (db *DB) CreatePolicy(ctx context.Context, p *Policy) (*Policy, error) {
	return db.CreatePolicyTx(ctx, nil, p)
}

func (db *DB) CreatePolicyTx(ctx context.Context, tx *sql.Tx, p *Policy) (*Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := p.Normalize()
	if err != nil {
		return nil, err
	}

	stmt := db.createPolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	n.ID = uuid.NewV4().String()

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.Repeat)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (db *DB) UpdatePolicy(ctx context.Context, p *Policy) error {
	return db.UpdatePolicyTx(ctx, nil, p)
}

func (db *DB) UpdatePolicyTx(ctx context.Context, tx *sql.Tx, p *Policy) error {
	err := validate.UUID("EscalationPolicyID", p.ID)
	if err != nil {
		return err
	}
	n, err := p.Normalize()
	if err != nil {
		return err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	stmt := db.updatePolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.Repeat)
	if err != nil {
		return err
	}

	db.logChange(ctx, nil, p.ID)

	return nil
}

func (db *DB) DeletePolicy(ctx context.Context, id string) error {
	return db.DeletePolicyTx(ctx, nil, id)
}

func (db *DB) DeletePolicyTx(ctx context.Context, tx *sql.Tx, id string) error {
	return db.DeleteManyPoliciesTx(ctx, tx, []string{id})
}

func (db *DB) DeleteManyPoliciesTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("EscalationPolicyID", ids, 50)
	if err != nil {
		return err
	}

	s := db.deletePolicy
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}
	_, err = s.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

func (db *DB) FindOnePolicy(ctx context.Context, id string) (*Policy, error) {
	return db.FindOnePolicyTx(ctx, nil, id)
}

func (db *DB) FindOnePolicyTx(ctx context.Context, tx *sql.Tx, id string) (*Policy, error) {
	err := validate.UUID("EscalationPolicyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := db.findOnePolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var p Policy
	err = row.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
	return &p, err
}

func (db *DB) FindOnePolicyForUpdateTx(ctx context.Context, tx *sql.Tx, id string) (*Policy, error) {
	err := validate.UUID("EscalationPolicyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := db.findOnePolicyForUpdate
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var p Policy
	err = row.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
	return &p, err
}

func (db *DB) FindAllPolicies(ctx context.Context) ([]Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	rows, err := db.findAllPolicies.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var p Policy
	policies := []Policy{}
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil

}

func (db *DB) FindAllPoliciesBySchedule(ctx context.Context, scheduleID string) ([]Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	rows, err := db.findAllPoliciesBySchedule.QueryContext(ctx, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var p Policy
	var policies []Policy
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
		if err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}

	return policies, nil
}

func (db *DB) FindOneStep(ctx context.Context, id string) (*Step, error) {
	return db.FindOneStepTx(ctx, nil, id)
}

func (db *DB) FindOneStepTx(ctx context.Context, tx *sql.Tx, id string) (*Step, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("EscalationPolicyStepID ", id)
	if err != nil {
		return nil, err
	}

	stmt := db.findOneStep
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var s Step
	err = row.Scan(&s.ID, &s.PolicyID, &s.DelayMinutes, &s.StepNumber)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (db *DB) FindOneStepForUpdateTx(ctx context.Context, tx *sql.Tx, id string) (*Step, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("EscalationPolicyStepID ", id)
	if err != nil {
		return nil, err
	}

	stmt := db.findOneStepForUpdate
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var s Step
	err = row.Scan(&s.ID, &s.PolicyID, &s.DelayMinutes, &s.StepNumber)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (db *DB) FindAllSteps(ctx context.Context, policyID string) ([]Step, error) {
	return db.FindAllStepsTx(ctx, nil, policyID)
}
func (db *DB) FindAllOnCallStepsForUserTx(ctx context.Context, tx *sql.Tx, userID string) ([]Step, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	stmt := db.findAllOnCallSteps
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Step
	for rows.Next() {
		var s Step
		err = rows.Scan(&s.ID, &s.PolicyID, &s.DelayMinutes, &s.StepNumber)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}
func (db *DB) FindAllStepsTx(ctx context.Context, tx *sql.Tx, policyID string) ([]Step, error) {
	err := validate.UUID("EscalationPolicyID", policyID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := db.findAllSteps
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	rows, err := stmt.QueryContext(ctx, policyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Step
	for rows.Next() {
		var s Step
		err = rows.Scan(&s.ID, &s.PolicyID, &s.DelayMinutes, &s.StepNumber)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

func (db *DB) CreateStep(ctx context.Context, s *Step) (*Step, error) {
	return db.CreateStepTx(ctx, nil, s)
}

func (db *DB) CreateStepTx(ctx context.Context, tx *sql.Tx, s *Step) (*Step, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := s.Normalize()
	if err != nil {
		return nil, err
	}

	stmt := db.createStep
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	n.ID = uuid.NewV4().String()

	err = stmt.QueryRowContext(ctx, n.ID, n.PolicyID, n.DelayMinutes).Scan(&n.StepNumber)
	if err != nil {
		return nil, err
	}

	db.logChange(ctx, tx, s.PolicyID)

	return n, nil
}

func (db *DB) UpdateStepNumberTx(ctx context.Context, tx *sql.Tx, stepID string, stepNumber int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.UUID("EscalationPolicyStepID", stepID)
	if err != nil {
		return err
	}

	numStmt := db.updateStepNumber
	if tx != nil {
		numStmt = tx.StmtContext(ctx, numStmt)
	}

	_, err = numStmt.ExecContext(ctx, stepID, stepNumber)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateStep(ctx context.Context, s *Step) error {
	return db.UpdateStepDelayTx(ctx, nil, s.ID, s.DelayMinutes)
}

func (db *DB) UpdateStepDelayTx(ctx context.Context, tx *sql.Tx, stepID string, stepDelay int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.UUID("EscalationPolicyStepID", stepID)
	if err != nil {
		return err
	}

	err = validate.Range("DelayMinutes", stepDelay, 1, 9000)
	if err != nil {
		return err
	}

	stmt := db.updateStepDelay
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, stepID, stepDelay)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) DeleteStep(ctx context.Context, id string) (string, error) {
	return db.DeleteStepTx(ctx, nil, id)
}

func (db *DB) DeleteStepTx(ctx context.Context, tx *sql.Tx, id string) (string, error) {
	err := validate.UUID("EscalationPolicyStepID", id)
	if err != nil {
		return "", err
	}

	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return "", err
	}
	s := db.deleteStep
	if tx != nil {
		s = tx.StmtContext(ctx, s)
	}
	row := s.QueryRowContext(ctx, id)
	var polID string
	err = row.Scan(&polID)
	if err != nil {
		return "", err
	}

	db.logChange(ctx, tx, polID)

	return polID, nil
}

func (db *DB) MoveStep(ctx context.Context, id string, newPos int) error {
	err := validate.Many(
		validate.UUID("EscalationPolicyStepID", id),
		validate.Range("NewPosition", newPos, 0, 9000),
	)
	if err != nil {
		return err
	}
	err = permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	var polID string
	err = db.moveStep.QueryRowContext(ctx, id, newPos).Scan(&polID)
	if err != nil {
		return err
	}
	db.logChange(ctx, nil, polID)

	return nil
}
func (db *DB) FindAllNotices(ctx context.Context, policyID string) ([]Notice, error) {

}
