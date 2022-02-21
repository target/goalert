package calsub

import (
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/validation/validate"
	"gorm.io/gorm"
)

// Subscription stores the information from user subscriptions
type Subscription struct {
	ID         string
	Name       string
	UserID     string
	ScheduleID string
	LastUpdate time.Time `gorm:"autoUpdateTime"`
	LastAccess time.Time
	Disabled   bool

	// Config provides necessary parameters CalendarSubscription Config (i.e. ReminderMinutes)
	Config SubscriptionConfig

	token string
}

func (Subscription) TableName() string { return "user_calendar_subscriptions" }

func (cs *Subscription) BeforeUpdate(db *gorm.DB) error {
	err := validate.Many(
		validate.Range("ReminderMinutes", len(cs.Config.ReminderMinutes), 0, 15),
		validate.IDName("Name", cs.Name),
		validate.UUID("ID", cs.ID),
	)
	if err != nil {
		return err
	}

	db.Statement.
		// if not the same user, they will get not-found
		Where("user_id = ?", permission.UserID(db.Statement.Context)).

		// limit updatable fields
		Select("name", "disabled", "config", "last_update")
	return nil
}

// Token returns the authorization token associated with this CalendarSubscription. It
// is only available when calling CreateTx.
func (cs Subscription) Token() string { return cs.token }

// Normalize will validate and produce a normalized CalendarSubscription struct.
func (cs Subscription) Normalize() (*Subscription, error) {
	if cs.ID == "" {
		cs.ID = uuid.New().String()
	}

	err := validate.Many(
		validate.Range("ReminderMinutes", len(cs.Config.ReminderMinutes), 0, 15),
		validate.IDName("Name", cs.Name),
		validate.UUID("ID", cs.ID),
		validate.UUID("UserID", cs.UserID),
	)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}
