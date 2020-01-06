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

func (a *CalendarSubscription) NotificationMinutes(ctx context.Context, obj *calendarsubscription.CalendarSubscription) ([]int, error) {
	var config calendarsubscription.Config
	err := json.Unmarshal(obj.Config, &config)
	obj.NotificationMinutes = config.NotificationMinutes
	return obj.NotificationMinutes, err
}
func (a *CalendarSubscription) ScheduleID(ctx context.Context, obj *calendarsubscription.CalendarSubscription) (string, error) {
	e := *obj
	return e.ScheduleID, nil
}
func (a *CalendarSubscription) Schedule(ctx context.Context, obj *calendarsubscription.CalendarSubscription) (*schedule.Schedule, error) {
	return a.ScheduleStore.FindOne(ctx, obj.ScheduleID)
}
func (a *CalendarSubscription) URL(ctx context.Context, obj *calendarsubscription.CalendarSubscription) (*string, error) {
	var err error
	return nil, err
}

func (q *Query) CalendarSubscription(ctx context.Context, id string) (*calendarsubscription.CalendarSubscription, error) {
	return q.CalendarSubscriptionStore.FindOne(ctx, id)
}
func (q *Query) CalendarSubscriptions(ctx context.Context) (subscription []calendarsubscription.CalendarSubscription, err error) {
	return q.CalendarSubscriptionStore.FindAll(ctx)
}

// todo: return url instead of bool once endpoint has been created
func (m *Mutation) CreateCalendarSubscription(ctx context.Context, input graphql2.CreateCalendarSubscriptionInput) (cs *calendarsubscription.CalendarSubscription, err error) {
	var config calendarsubscription.Config
	var configJson []byte
	if input.NotificationMinutes != nil {
		config.NotificationMinutes = input.NotificationMinutes
		configJson, err = json.Marshal(config)
		if err != nil {
			return nil, err
		}
	}
	cs = &calendarsubscription.CalendarSubscription{
		Name:       input.Name,
		ScheduleID: input.ScheduleID,
		Config:     configJson,
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

func (m *Mutation) UpdateCalendarSubscription(ctx context.Context, input graphql2.UpdateCalendarSubscriptionInput) (res bool, err error) {
	var config calendarsubscription.Config
	var configJson []byte
	if input.NotificationMinutes != nil {
		config.NotificationMinutes = input.NotificationMinutes
		configJson, err = json.Marshal(config)
		if err != nil {
			return res, err
		}
	}

	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cs := &calendarsubscription.CalendarSubscription{
			ID:       input.ID,
			Name:     *input.Name,
			Disabled: *input.Disabled,
			Config:   configJson,
		}

		err := m.CalendarSubscriptionStore.UpdateSubscriptionTx(ctx, tx, cs)
		if err != nil {
			return err
		}

		return err
	})

	return err == nil, err
}

