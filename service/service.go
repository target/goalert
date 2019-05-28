package service

import "github.com/target/goalert/validation/validate"

type Service struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	EscalationPolicyID string `json:"escalation_policy_id"`

	epName         string
	isUserFavorite bool
}

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
	err := validate.Many(
		validate.IDName("Name", s.Name),
		validate.Text("Description", s.Description, 1, 255),
		validate.UUID("EscalationPolicyID", s.EscalationPolicyID),
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
