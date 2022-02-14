package graphqlapp

import (
	"context"
	"net/url"

	"github.com/target/goalert/calsub"
	"github.com/target/goalert/config"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/permission"
	"github.com/target/goalert/schedule"
	"github.com/target/goalert/util/sqlutil"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserCalendarSubscription App

func (a *App) UserCalendarSubscription() graphql2.UserCalendarSubscriptionResolver {
	return (*UserCalendarSubscription)(a)
}

func (a *UserCalendarSubscription) ReminderMinutes(ctx context.Context, obj *calsub.Subscription) ([]int, error) {
	return obj.Config.ReminderMinutes, nil
}
func (a *UserCalendarSubscription) Schedule(ctx context.Context, obj *calsub.Subscription) (*schedule.Schedule, error) {
	return a.ScheduleStore.FindOne(ctx, obj.ScheduleID)
}
func (a *UserCalendarSubscription) URL(ctx context.Context, obj *calsub.Subscription) (*string, error) {
	tok, err := a.CalSubStore.SignToken(ctx, obj)
	if err != nil {
		return nil, err
	}
	if tok == "" {
		return nil, nil
	}

	v := make(url.Values)
	v.Set("token", tok)

	cfg := config.FromContext(ctx)
	callback := cfg.CallbackURL("/api/v2/calendar", v)
	return &callback, nil
}

func (q *Query) UserCalendarSubscription(ctx context.Context, id string) (*calsub.Subscription, error) {
	var cs calsub.Subscription
	err := sqlutil.FromContext(ctx).Where("id = ?", id).Take(&cs).Error
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// todo: return UserCalendarSubscription with generated url once endpoint has been created
func (m *Mutation) CreateUserCalendarSubscription(ctx context.Context, input graphql2.CreateUserCalendarSubscriptionInput) (cs *calsub.Subscription, err error) {
	cs = &calsub.Subscription{
		Name:       input.Name,
		ScheduleID: input.ScheduleID,
		UserID:     permission.UserID(ctx),
	}
	if input.Disabled != nil {
		cs.Disabled = *input.Disabled
	}
	cs.Config.ReminderMinutes = input.ReminderMinutes

	err = sqlutil.FromContext(ctx).Create(cs).Error
	if err != nil {
		return nil, err
	}

	return cs, nil
}

func (m *Mutation) UpdateUserCalendarSubscription(ctx context.Context, input graphql2.UpdateUserCalendarSubscriptionInput) (bool, error) {
	err := sqlutil.Transaction(ctx, func(ctx context.Context, db *gorm.DB) error {
		var cs calsub.Subscription
		err := db.Where("id = ?", input.ID).Clauses(clause.Locking{Strength: "FOR UPDATE"}).Take(&cs).Error
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}
		if input.Name != nil {
			cs.Name = *input.Name
		}
		if input.Disabled != nil {
			cs.Disabled = *input.Disabled
		}
		if input.ReminderMinutes != nil {
			cs.Config.ReminderMinutes = input.ReminderMinutes
		}
		return db.Save(&cs).Error
	})

	return err == nil, err
}
