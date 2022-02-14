package calsub

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/auth/authtoken"
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
	LastAccess time.Time
	Disabled   bool

	// Config provides necessary parameters CalendarSubscription Config (i.e. ReminderMinutes)
	Config SubscriptionConfig

	token *authtoken.Token
}

// SubscriptionConfig is the configuration for a calendar subscription.
type SubscriptionConfig struct {
	ReminderMinutes []int
}

var (
	_ = driver.Valuer(&SubscriptionConfig{})
	_ = sql.Scanner(&SubscriptionConfig{})
)

func (scfg *SubscriptionConfig) Value() (driver.Value, error) {
	if scfg == nil {
		return nil, nil
	}
	data, err := json.Marshal(scfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}
func (scfg *SubscriptionConfig) Scan(v interface{}) error {
	switch v := v.(type) {
	case []byte:
		return json.Unmarshal(v, scfg)
	case string:
		return json.Unmarshal([]byte(v), scfg)
	default:
		return fmt.Errorf("unsupported type %T", v)
	}
}

func (cs *Subscription) BeforeCreate(db *gorm.DB) error {
	err := permission.LimitCheckAny(db.Statement.Context, permission.MatchUser(cs.UserID))
	if err != nil {
		return err
	}

	err = validate.Many(
		validate.Range("ReminderMinutes", len(cs.Config.ReminderMinutes), 0, 15),
		validate.IDName("Name", cs.Name),
		validate.UUID("ID", cs.ID),
		validate.UUID("UserID", cs.UserID),
		validate.UUID("ScheduleID", cs.ScheduleID),
	)
	if err != nil {
		return err
	}

	var now time.Time
	err = db.Raw("select now()").Scan(&now).Error
	if err != nil {
		return nil
	}

	cs.token = &authtoken.Token{
		Type:      authtoken.TypeCalSub,
		Version:   2,
		CreatedAt: now,
		ID:        uuid.New(),
	}
	cs.ID = cs.token.ID.String()
	return nil
}

func (cs *Subscription) BeforeUpdate(db *gorm.DB) error {
	// if not the same user, they will get not-found
	db.Statement.Where("user_id = ?", permission.UserID(db.Statement.Context))

	err := validate.Many(
		validate.Range("ReminderMinutes", len(cs.Config.ReminderMinutes), 0, 15),
		validate.IDName("Name", cs.Name),
		validate.UUID("ID", cs.ID),
	)
	if err != nil {
		return err
	}

	db.Statement.Select("name", "disabled", "config")
	db.Statement.UpdateColumn("last_update", gorm.Expr("now()"))
	return nil
}

func (cs *Subscription) BeforeDelete(db *gorm.DB) error {
	db.Statement.Where("user_id = ?", permission.UserID(db.Statement.Context))
	return nil
}
