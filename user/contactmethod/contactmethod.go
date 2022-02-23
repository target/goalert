package contactmethod

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/validation/validate"
)

// ContactMethod stores the information for contacting a user.
type ContactMethod struct {
	ID       string
	Name     string
	Type     Type
	Value    string
	Disabled bool
	UserID   string

	lastTestVerifyAt sql.NullTime
}

func (ContactMethod) TableName() string { return "user_contact_methods" }

// LastTestVerifyAt will return the timestamp of the last test/verify request.
func (c ContactMethod) LastTestVerifyAt() time.Time { return c.lastTestVerifyAt.Time }

// Normalize will validate and 'normalize' the ContactMethod -- such as making email lower-case
// and setting carrier to "" (for non-phone types).
func (c ContactMethod) Normalize() (*ContactMethod, error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	err := validate.Many(
		validate.UUID("ID", c.ID),
		validate.IDName("Name", c.Name),
		validate.OneOf("Type", c.Type, TypeSMS, TypeVoice, TypeEmail, TypePush, TypeWebhook),
	)

	switch c.Type {
	case TypeSMS, TypeVoice:
		err = validate.Many(err, validate.Phone("Value", c.Value))
	case TypeEmail:
		err = validate.Many(err, validate.Email("Value", c.Value))
	case TypeWebhook:
		err = validate.Many(err, validate.AbsoluteURL("Value", c.Value))
	case TypePush:
		c.Value = ""
	}

	if err != nil {
		return nil, err
	}

	return &c, nil
}
