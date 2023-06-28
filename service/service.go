package service

import (
	"time"

	"github.com/target/goalert/validation/validate"
)

type Service struct {
	ID                   string
	Name                 string
	Description          string
	EscalationPolicyID   string
	MaintenanceExpiresAt time.Time

	epName         string
	isUserFavorite bool
}

const MaxDetailsLength = 6 * 1024 // 6KiB

func (s Service) EscalationPolicyName() string {
	return s.epName
}

// IsUserFavorite returns a boolean value based on if the service is a favorite of the user or not.
func (s Service) IsUserFavorite() bool {
	return s.isUserFavorite
}

// Normalize will validate and 'normalize' the ContactMethod -- such as making email lower-case
// and setting carrier to "" (for non-phone types).
func (s Service) Normalize() (*Service, error) {
	dur := time.Until(s.MaintenanceExpiresAt)

	if dur <= 0 {
		dur = 0
		s.MaintenanceExpiresAt = time.Time{}
	}

	err := validate.Many(
		validate.IDName("Name", s.Name),
		validate.Text("Description", s.Description, 1, MaxDetailsLength),
		validate.UUID("EscalationPolicyID", s.EscalationPolicyID),
		validate.Duration("MaintenanceExpiresAt", dur, 0, 8*time.Hour),
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
