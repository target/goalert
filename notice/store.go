package notice

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/target/goalert/config"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// Store allows identifying notices for various targets.
type Store struct {
	findServicesByPolicyID    *sql.Stmt
	findPolicyDurationMinutes *sql.Stmt
}

// NewStore creates a new DB and prepares all necessary SQL statements.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
		findServicesByPolicyID: p.P(`
			SELECT COUNT(*)
			FROM services
			WHERE escalation_policy_id = $1
		`),
		findPolicyDurationMinutes: p.P(`
			SELECT SUM(s.delay*e.repeat)
		    FROM escalation_policy_steps s join escalation_policies e
		    ON s.escalation_policy_id= e.id 
			WHERE s.escalation_policy_id=$1
		`),
	}, p.Err
}

// FindAllPolicyNotices sets a notice for a Policy if it is not assigned to any services.
func (s *Store) FindAllPolicyNotices(ctx context.Context, policyID string) ([]Notice, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("EscalationPolicyStepID", policyID)
	if err != nil {
		return nil, err
	}

	var numServices int
	err = s.findServicesByPolicyID.QueryRowContext(ctx, policyID).Scan(&numServices)
	if err != nil {
		return nil, err
	}

	var notices []Notice
	if numServices == 0 {
		notices = append(notices, Notice{
			Message: "Not assigned to a service",
			Details: "To receive alerts for this configuration, assign this escalation policy to its relevant service.",
		})
	}
	type PolicyDuration struct {
		DelaySum int
	}
	cfg := config.FromContext(ctx)
	var policyDuration PolicyDuration
	err = s.findPolicyDurationMinutes.QueryRowContext(ctx, policyID).Scan(&policyDuration.DelaySum)
	if err != nil {
		return nil, err
	}

	autoCloseDays := cfg.Maintenance.AlertAutoCloseDays

	if autoCloseDays > 0 && policyDuration.DelaySum/(24*60) >= autoCloseDays {
		notices = append(notices, Notice{
			Message: "Auto Closure of uncknowledged/stale alerts",
			Details: "Alerts using this policy will be automatically closed after " + strconv.Itoa(autoCloseDays) + " days.",
		})
	}

	return notices, nil
}
