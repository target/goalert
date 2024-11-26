package notice

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/target/goalert/config"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// Store allows identifying notices for various targets.
type Store struct {
	db                        *sql.DB
	findServicesByPolicyID    *sql.Stmt
	findPolicyDurationMinutes *sql.Stmt
}

// NewStore creates a new DB and prepares all necessary SQL statements.
func NewStore(ctx context.Context, db *sql.DB) (*Store, error) {
	p := &util.Prepare{DB: db, Ctx: ctx}
	return &Store{
		db: db,
		findServicesByPolicyID: p.P(`
			SELECT COUNT(*)
			FROM services
			WHERE escalation_policy_id = $1
		`),
		findPolicyDurationMinutes: p.P(`
			SELECT coalesce(SUM(s.delay*(e.repeat+1)), 0)
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
	var mins int
	err = s.findPolicyDurationMinutes.QueryRowContext(ctx, policyID).Scan(&mins)
	if err != nil {
		return nil, err
	}

	cfg := config.FromContext(ctx)
	days := cfg.Maintenance.AlertAutoCloseDays

	if days > 0 && mins/(24*60) >= days {
		notices = append(notices, Notice{
			Message: "Auto-closure of unacknowledged alerts",
			Details: fmt.Sprintf("Alerts using this policy will be automatically closed after %d day(s).", days),
		})
	}

	return notices, nil
}

// FindAllServiceNotices returns any relevant notices for the given service. Currently returns
// notices pertaining to the system limit for unacknowledged alerts.
func (s *Store) FindAllServiceNotices(ctx context.Context, serviceID string) ([]Notice, error) {
	err := permission.LimitCheckAny(ctx, permission.User)
	if err != nil {
		return nil, err
	}

	err = validate.UUID("serviceID", serviceID)
	if err != nil {
		return nil, err
	}

	res, err := gadb.NewCompat(s.db).NoticeUnackedAlertsByService(ctx, uuid.MustParse(serviceID))
	if err != nil {
		return nil, err
	}

	count := float32(res.Count)
	limit := float32(res.Max)
	details := fmt.Sprintf("New alerts will be rejected while at or above the limit (%v), acknowledge or close alerts to resolve.", count)

	// hit system limit notice
	if count >= limit {
		return []Notice{
			{
				Type:    TypeError,
				Message: "Unacknowledged alert limit reached",
				Details: details,
			},
		}, nil
	}

	// nearing system limit notice
	if count/limit > 0.75 {
		return []Notice{
			{
				Type:    TypeWarning,
				Message: "Near unacknowledged alert limit",
				Details: details,
			},
		}, nil
	}

	return nil, nil
}
