package user

import (
	"github.com/target/goalert/validation/validate"
)

// An AuthSubject contains information about the auth provider and subject ID for a particular user.
type AuthSubject struct {
	// ProviderID is the ID for the provider of the user.
	ProviderID string

	// SubjectID is the ID for the subject of the user.
	SubjectID string

	// UserID is the ID of the user.
	UserID string
}

// Normalize will validate and produce a normalized AuthSubject struct.
func (a AuthSubject) Normalize() (*AuthSubject, error) {
	err := validate.Many(
		validate.SubjectID("SubjectID", a.SubjectID),
		validate.SubjectID("ProviderID", a.ProviderID),
		validate.UUID("UserID", a.UserID),
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}
