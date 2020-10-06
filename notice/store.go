package notice

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/twilio"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/util"
	"github.com/target/goalert/validation/validate"
)

// Store allows identifying notices for various targets.
type Store struct {
	findServicesByPolicyID *sql.Stmt

	tw *twilio.Config
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

// SetTwilioConfig is used to set Twilio Config & client.
func (s *Store) SetTwilioConfig(tw *twilio.Config) { s.tw = tw }

// FindAllConfigNotices returns notices for potential config problems.
//
// SetTwilioConfig must be called before calling FindAllConfigNotices.
func (s *Store) FindAllConfigNotices(ctx context.Context) ([]Notice, error) {
	err := permission.LimitCheckAny(ctx, permission.Admin)
	if err != nil {
		return nil, err
	}

	cfg := config.FromContext(ctx)
	if !cfg.Twilio.Enable {
		return nil, nil
	}

	var notices []Notice

	phoneN, err := s.checkPhoneCallback(ctx, cfg.Twilio.FromNumber)
	if err != nil {
		return nil, err
	}
	notices = append(notices, phoneN...)

	for _, number := range cfg.TwilioSMSFromNumbers() {
		phoneN, err = s.checkPhoneCallback(ctx, number)
		if err != nil {
			return nil, err
		}
		notices = append(notices, phoneN...)
	}

	return notices, nil
}
func (s *Store) checkPhoneCallback(ctx context.Context, number string) ([]Notice, error) {
	cfg := config.FromContext(ctx)
	var notices []Notice

	phone, err := s.tw.PhoneNumberConfig(ctx, number)
	if err != nil {
		return nil, fmt.Errorf("check Twilio number config: %w", err)
	}
	if phone == nil {
		return []Notice{{
			Type:    TypeError,
			Message: fmt.Sprintf("Twilio: %s does not exist on the configured account", number),
			Details: "SMS and voice messages will fail to send from this number",
		}}, nil
	}
	if !phone.Capabilities.SMS {
		notices = append(notices, Notice{
			Type:    TypeWarning,
			Message: fmt.Sprintf("Twilio: %s does not support SMS", number),
			Details: "SMS messages will fail to send from this number",
		})
	}
	if !phone.Capabilities.Voice {
		notices = append(notices, Notice{
			Type:    TypeWarning,
			Message: fmt.Sprintf("Twilio: %s does not support voice", number),
		})
	}

	var problems []string
	if phone.SMSMethod != "POST" || phone.SMSURL != cfg.CallbackURL("/api/v2/twilio/message") {
		problems = append(problems, "SMS webhook method/URL is incorrect, SMS replies may fail")
	}
	if phone.VoiceMethod != "POST" || phone.VoiceURL != cfg.CallbackURL("/api/v2/twilio/call") {
		problems = append(problems, "voice webhook method/URL is incorrect, incoming calls may fail")
	}

	if len(problems) > 0 {
		notices = append(notices, Notice{
			Type:    TypeWarning,
			Message: fmt.Sprintf("Twilio: %s has incorrect configuration", number),
			Details: strings.Join(problems, "\n"),
		})
	}

	return notices, nil
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
