package escalation

import (
	"context"
	"database/sql"

	"github.com/target/goalert/alert/alertlog"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/notification/slack"
	"github.com/target/goalert/notificationchannel"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/util/log"
	"github.com/target/goalert/util/sqlutil"
	"github.com/target/goalert/validation/validate"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type Config struct {
	NCStore         *notificationchannel.Store
	Registry        *nfydest.Registry
	LogStore        *alertlog.Store
	SlackLookupFunc func(ctx context.Context, channelID string) (*slack.Channel, error)
}

type Store struct {
	db *sql.DB

	log     *alertlog.Store
	ncStore *notificationchannel.Store
	reg     *nfydest.Registry
	slackFn func(ctx context.Context, channelID string) (*slack.Channel, error)

	findOnePolicy          *sql.Stmt
	findOnePolicyForUpdate *sql.Stmt
	findManyPolicies       *sql.Stmt

	findAllPoliciesBySchedule *sql.Stmt
	createPolicy              *sql.Stmt
	updatePolicy              *sql.Stmt
	deletePolicy              *sql.Stmt

	findOneStepForUpdate *sql.Stmt
	findAllSteps         *sql.Stmt
	findAllOnCallSteps   *sql.Stmt
	createStep           *sql.Stmt
	updateStepDelay      *sql.Stmt
	updateStepNumber     *sql.Stmt
	deleteStep           *sql.Stmt
}

func NewStore(ctx context.Context, db *sql.DB, cfg Config) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}

	return &Store{
		db:      db,
		log:     cfg.LogStore,
		slackFn: cfg.SlackLookupFunc,
		ncStore: cfg.NCStore,
		reg:     cfg.Registry,

		findOnePolicy: p.P(`
			SELECT
				e.id,
				e.name,
				e.description,
				e.repeat,
				fav is distinct from null
			FROM
				escalation_policies e
			LEFT JOIN user_favorites fav ON
				fav.tgt_escalation_policy_id = e.id AND fav.user_id = $2
			WHERE e.id = $1
		`),
		findOnePolicyForUpdate: p.P(`SELECT id, name, description, repeat FROM escalation_policies WHERE id = $1 FOR UPDATE`),
		findManyPolicies: p.P(`
            SELECT
                e.id,
                e.name,
                e.description,
                e.repeat,
                fav is distinct from null
            FROM
                escalation_policies e
            LEFT JOIN user_favorites fav ON
                fav.tgt_escalation_policy_id = e.id AND fav.user_id = $2
            WHERE e.id = any($1)
        `),
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
	}, p.Err
}

func (s *Store) logChange(ctx context.Context, tx *sql.Tx, policyID string) {
	err := s.log.LogEPTx(ctx, tx, policyID, alertlog.TypePolicyUpdated, nil)
	if err != nil {
		log.Log(ctx, errors.Wrap(err, "append alertlog (escalation policy update)"))
	}
}

// FindManyPolicies returns escalation policies for the given IDs.
func (s *Store) FindManyPolicies(ctx context.Context, ids []string) ([]Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.ManyUUID("EscalationPolicyID", ids, 200)
	if err != nil {
		return nil, err
	}

	rows, err := s.findManyPolicies.QueryContext(ctx, sqlutil.UUIDArray(ids), permission.UserNullUUID(ctx))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Policy
	var p Policy
	for rows.Next() {
		err = rows.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat, &p.isUserFavorite)
		if err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	return result, nil
}

// CreatePolicyTx creates a new escalation policy in the database.
func (s *Store) CreatePolicyTx(ctx context.Context, tx *sql.Tx, p *Policy) (*Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := p.Normalize()
	if err != nil {
		return nil, err
	}

	stmt := s.createPolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	n.ID = uuid.New().String()

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.Repeat)
	if err != nil {
		return nil, err
	}
	return n, nil
}

// UpdatePolicyTx will update a single escalation policy.
func (s *Store) UpdatePolicyTx(ctx context.Context, tx *sql.Tx, p *Policy) error {
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

	stmt := s.updatePolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, n.ID, n.Name, n.Description, n.Repeat)
	if err != nil {
		return err
	}

	s.logChange(ctx, nil, p.ID)

	return nil
}

// DeleteManyPoliciesTx deletes multiple policies in a single transaction.
func (s *Store) DeleteManyPoliciesTx(ctx context.Context, tx *sql.Tx, ids []string) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}
	err = validate.ManyUUID("EscalationPolicyID", ids, 50)
	if err != nil {
		return err
	}

	stmt := s.deletePolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	_, err = stmt.ExecContext(ctx, sqlutil.UUIDArray(ids))
	return err
}

// FindOnePolicyTx returns a policy by ID.
func (s *Store) FindOnePolicyTx(ctx context.Context, tx *sql.Tx, id string) (*Policy, error) {
	err := validate.UUID("EscalationPolicyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := s.findOnePolicy
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var p Policy
	err = row.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
	return &p, err
}

// FindOnePolicyForUpdateTx returns a single policy locked to the tx for updating.
func (s *Store) FindOnePolicyForUpdateTx(ctx context.Context, tx *sql.Tx, id string) (*Policy, error) {
	err := validate.UUID("EscalationPolicyID", id)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := s.findOnePolicyForUpdate
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var p Policy
	err = row.Scan(&p.ID, &p.Name, &p.Description, &p.Repeat)
	return &p, err
}

// FindAllPoliciesBySchedule will return all policies that have the given schedule assigned to them.
func (s *Store) FindAllPoliciesBySchedule(ctx context.Context, scheduleID string) ([]Policy, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("ScheduleID", scheduleID)
	if err != nil {
		return nil, err
	}
	rows, err := s.findAllPoliciesBySchedule.QueryContext(ctx, scheduleID)
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

// FindOneStepForUpdateTx returns a step locked within the tx for update.
func (s *Store) FindOneStepForUpdateTx(ctx context.Context, tx *sql.Tx, id string) (*Step, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("EscalationPolicyStepID ", id)
	if err != nil {
		return nil, err
	}

	stmt := s.findOneStepForUpdate
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	row := stmt.QueryRowContext(ctx, id)
	var st Step
	err = row.Scan(&st.ID, &st.PolicyID, &st.DelayMinutes, &st.StepNumber)
	if err != nil {
		return nil, err
	}

	return &st, nil
}

func (s *Store) FindAllSteps(ctx context.Context, policyID string) ([]Step, error) {
	return s.FindAllStepsTx(ctx, nil, policyID)
}

// FindAllOnCallStepsForUserTx returns all steps a user is currently on-call for.
func (s *Store) FindAllOnCallStepsForUserTx(ctx context.Context, tx *sql.Tx, userID string) ([]Step, error) {
	err := permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}
	err = validate.UUID("UserID", userID)
	if err != nil {
		return nil, err
	}

	stmt := s.findAllOnCallSteps
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

// FindAllStepsTx returns all steps for a policy.
func (s *Store) FindAllStepsTx(ctx context.Context, tx *sql.Tx, policyID string) ([]Step, error) {
	err := validate.UUID("EscalationPolicyID", policyID)
	if err != nil {
		return nil, err
	}

	err = permission.LimitCheckAny(ctx, permission.All)
	if err != nil {
		return nil, err
	}

	stmt := s.findAllSteps
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

// CreateStepTx adds a step to an escalation policy.
func (s *Store) CreateStepTx(ctx context.Context, tx *sql.Tx, st *Step) (*Step, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return nil, err
	}

	n, err := st.Normalize()
	if err != nil {
		return nil, err
	}

	stmt := s.createStep
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	n.ID = uuid.New()

	err = stmt.QueryRowContext(ctx, n.ID, n.PolicyID, n.DelayMinutes).Scan(&n.StepNumber)
	if err != nil {
		return nil, err
	}

	s.logChange(ctx, tx, st.PolicyID)
	return n, nil
}

// UpdateStepNumberTx updates the step number for a step.
func (s *Store) UpdateStepNumberTx(ctx context.Context, tx *sql.Tx, stepID uuid.UUID, stepNumber int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	numStmt := s.updateStepNumber
	if tx != nil {
		numStmt = tx.StmtContext(ctx, numStmt)
	}

	_, err = numStmt.ExecContext(ctx, stepID, stepNumber)
	if err != nil {
		return err
	}

	return nil
}

// UpdateStepDelayTx updates the delay for a step.
func (s *Store) UpdateStepDelayTx(ctx context.Context, tx *sql.Tx, stepID uuid.UUID, stepDelay int) error {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return err
	}

	err = validate.Range("DelayMinutes", stepDelay, 1, 9000)
	if err != nil {
		return err
	}

	stmt := s.updateStepDelay
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}

	_, err = stmt.ExecContext(ctx, stepID, stepDelay)
	if err != nil {
		return err
	}

	return nil
}

// DeleteStepTx deletes a step from an escalation policy.
func (s *Store) DeleteStepTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (string, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin, permission.User)
	if err != nil {
		return "", err
	}
	stmt := s.deleteStep
	if tx != nil {
		stmt = tx.StmtContext(ctx, stmt)
	}
	row := stmt.QueryRowContext(ctx, id)
	var polID string
	err = row.Scan(&polID)
	if err != nil {
		return "", err
	}

	s.logChange(ctx, tx, polID)

	return polID, nil
}
