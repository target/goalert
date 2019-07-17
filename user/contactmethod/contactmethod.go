package contactmethod

import (
	uuid "github.com/satori/go.uuid"
	"github.com/target/goalert/validation/validate"
)

// ContactMethod stores the information for contacting a user.
type ContactMethod struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     Type   `json:"type"`
	Value    string `json:"value"`
	Disabled bool   `json:"disabled"`
	UserID   string `json:"-"`
}

// Normalize will validate and 'normalize' the ContactMethod -- such as making email lower-case
// and setting carrier to "" (for non-phone types).
func (c ContactMethod) Normalize() (*ContactMethod, error) {
	if c.ID == "" {
		c.ID = uuid.NewV4().String()
	}
	err := validate.Many(
		validate.UUID("ID", c.ID),
		validate.IDName("Name", c.Name),
		validate.OneOf("Type", c.Type, TypeSMS, TypeVoice, TypeEmail, TypePush),
	)

	switch c.Type {
	case TypeSMS, TypeVoice:
		err = validate.Many(err, validate.Phone("Value", c.Value))
	case TypeEmail:
		err = validate.Many(err, validate.Email("Value", c.Value))
	case TypePush:
		c.Value = ""
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}
