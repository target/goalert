package notice

import (
	"context"
	"database/sql"

	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// Store allows identifying notices for various targets.
type Store struct {
	findServicesByPolicyID *sql.Stmt
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

	return notices, nil
}
