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
	Type     Type
	Value    string
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

	if c.Dest.Type == "" {
		// Set the destination type based on the contact method type.
		//
		// These are hard-coded for compatibility until refactor is complete, since otherwise we'd have import cycles.
		switch c.Type {
		case TypeSMS: // twilio.DestTypeSMS & twilio.FieldPhoneNumber
			c.Dest = gadb.NewDestV1("builtin-twilio-sms", "phone_number", c.Value)
		case TypeVoice: // twilio.DestTypeVoice & twilio.FieldPhoneNumber
			c.Dest = gadb.NewDestV1("builtin-twilio-voice", "phone_number", c.Value)
		case TypeEmail: // email.DestTypeEmail & email.FieldEmailAddress
			c.Dest = gadb.NewDestV1("builtin-email", "email_address", c.Value)
		case TypeWebhook: // webhook.DestTypeWebhook & webhook.FieldWebhookURL
			c.Dest = gadb.NewDestV1("builtin-webhook", "webhook_url", c.Value)
		case TypeSlackDM: // slack.DestTypeSlackDirectMessage & slack.FieldSlackUserID
			c.Dest = gadb.NewDestV1("builtin-slack-dm", "slack_user_id", c.Value)
		}
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
