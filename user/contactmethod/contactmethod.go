package contactmethod

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
	"github.com/target/goalert/validation/validate"
)

// ContactMethod stores the information for contacting a user.
type ContactMethod struct {
	ID       uuid.UUID
	Name     string
	Dest     gadb.DestV1
	Disabled bool
	UserID   string
	Pending  bool

	StatusUpdates bool

	lastTestVerifyAt sql.NullTime
}

// LastTestVerifyAt will return the timestamp of the last test/verify request.
func (c ContactMethod) LastTestVerifyAt() time.Time { return c.lastTestVerifyAt.Time }

// Normalize will validate and 'normalize' the ContactMethod -- such as making email lower-case
// and setting carrier to "" (for non-phone types).
func (c ContactMethod) Normalize(ctx context.Context, reg *nfydest.Registry) (*ContactMethod, error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	err := validate.IDName("Name", c.Name)
	if err != nil {
		return nil, err
	}

	info, err := reg.TypeInfo(ctx, c.Dest.Type)
	if err != nil {
		return nil, err
	}
	if !info.IsContactMethod() {
		return nil, validation.NewFieldError("Dest.Type", "invalid destination type")
	}

	if !info.SupportsStatusUpdates {
		c.StatusUpdates = false
	} else if info.StatusUpdatesRequired {
		c.StatusUpdates = true
	}

	err = reg.ValidateDest(ctx, c.Dest)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
