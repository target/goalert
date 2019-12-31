package graphqlapp

import (
	"context"
	"database/sql"
	"github.com/target/goalert/calendarsubscription"
	"github.com/target/goalert/graphql2"
	"github.com/target/goalert/search"
)

func (q *Query) CalendarSubscription(ctx context.Context, id string) (*calendarsubscription.CalendarSubscription, error) {
	return q.CalendarSubscriptionStore.FindOne(ctx, id)
}

// todo: return url instead of bool once endpoint has been created
func (m *Mutation) CreateCalendarSubscription(ctx context.Context, input graphql2.CreateCalendarSubscriptionInput) (res bool, err error) {
	err = withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cs := &calendarsubscription.CalendarSubscription{
			Name: input.Name,
		}

		res, err = m.CalendarSubscriptionStore.CreateSubscriptionTx(ctx, tx, cs)
		if err != nil {
			return err
		}

		// todo: gen url for user

		return err
	})

	return res, err
}

func (m *Mutation) UpdateCalendarSubscription(ctx context.Context, input graphql2.UpdateCalendarSubscriptionInput) (bool, error) {
	err := withContextTx(ctx, m.DB, func(ctx context.Context, tx *sql.Tx) error {
		cs := &calendarsubscription.CalendarSubscription{
			ID:       input.ID,
			Name:     *input.Name,
			Disabled: *input.Disabled,
		}

		err := m.CalendarSubscriptionStore.UpdateSubscriptionTx(ctx, tx, cs)
		if err != nil {
			return err
		}

		return err
	})

	return err == nil, err
}
func (q *Query) CalendarSubscriptions(ctx context.Context, opts *graphql2.CalendarSubscriptionSearchOptions) (conn *graphql2.CalendarSubscriptionConnection, err error) {
	if opts == nil {
		opts = &graphql2.CalendarSubscriptionSearchOptions{}
	}
	var searchOpts calendarsubscription.SearchOptions
	if opts.Search != nil {
		searchOpts.Search = *opts.Search
	}
	searchOpts.Omit = opts.Omit
	if opts.After != nil && *opts.After != "" {
		err = search.ParseCursor(*opts.After, &searchOpts)
		if err != nil {
			return nil, err
		}
	}
	if opts.First != nil {
		searchOpts.Limit = *opts.First
	}
	if searchOpts.Limit == 0 {
		searchOpts.Limit = 15
	}

	searchOpts.Limit++
	subs, err := q.CalendarSubscriptionStore.Search(ctx, &searchOpts)
	if err != nil {
		return nil, err
	}

	conn = new(graphql2.CalendarSubscriptionConnection)
	conn.PageInfo = &graphql2.PageInfo{}
	if len(subs) == searchOpts.Limit {
		subs = subs[:len(subs)-1]
		conn.PageInfo.HasNextPage = true
	}
	if len(subs) > 0 {
		last := subs[len(subs)-1]
		searchOpts.After.Name = last.Name

		cur, err := search.Cursor(searchOpts)
		if err != nil {
			return conn, err
		}
		conn.PageInfo.EndCursor = &cur
	}
	conn.Nodes = subs
	return conn, err

}
