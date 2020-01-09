package graphqlapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/target/goalert/calendarsubscription"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/schedule"
)

type CalendarSubscription App

func (a *App) CalendarSubscription() graphql2.CalendarSubscriptionResolver {
	return (*CalendarSubscription)(a)
}

func (a *CalendarSubscription) ReminderMinutes(ctx context.Context, obj *calendarsubscription.CalendarSubscription) ([]int, error) {
	var config calendarsubscription.Config
	err := json.Unmarshal(obj.Config, &config)
	obj.ReminderMinutes = config.ReminderMinutes
	return obj.ReminderMinutes, err
}
func (a *CalendarSubscription) Schedule(ctx context.Context, obj *calendarsubscription.CalendarSubscription) (*schedule.Schedule, error) {
	return a.ScheduleStore.FindOne(ctx, obj.ScheduleID)
}

func (q *Query) UserCalendarSubscription(ctx context.Context, id string) (*calendarsubscription.CalendarSubscription, error) {
	return q.CalendarSubscriptionStore.FindOne(ctx, id)
}

// todo: return calendarsubscription with generated url once endpoint has been created
func (m *Mutation) CreateUserCalendarSubscription(ctx context.Context, input graphql2.CreateUserCalendarSubscriptionInput) (cs *calendarsubscription.CalendarSubscription, err error) {
	var config calendarsubscription.Config
	var configJson []byte
	if input.ReminderMinutes != nil {
		config.ReminderMinutes = input.ReminderMinutes
		configJson, err = json.Marshal(config)
		if err != nil {
			return nil, err
		}
	}
	cs = &calendarsubscription.CalendarSubscription{
		Name:            input.Name,
		ScheduleID:      input.ScheduleID,
		Config:          configJson,
		ReminderMinutes: input.ReminderMinutes,
		Disabled:        *input.Disabled,
	}
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		var err error
		cs, err = m.CalendarSubscriptionStore.CreateSubscriptionTx(ctx, tx, cs)
		if err != nil {
			return err
		}

		// todo: gen url for user

		return err
	})

	return cs, err
}

func (m *Mutation) UpdateUserCalendarSubscription(ctx context.Context, input graphql2.UpdateUserCalendarSubscriptionInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cs, err := m.CalendarSubscriptionStore.FindOneForUpdateTx(ctx, tx, input.ID)
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
			var config calendarsubscription.Config
			var configJson []byte
			config.ReminderMinutes = input.ReminderMinutes
			configJson, err := json.Marshal(config)
			if err != nil {
				return err
			}
			cs.Config = configJson
		}
		return m.CalendarSubscriptionStore.UpdateTx(ctx, tx, cs)
	})
	return err == nil, err
}
