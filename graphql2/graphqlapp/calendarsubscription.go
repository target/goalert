package graphqlapp

import (
	"context"
	"github.com/target/goalert/calendarsubscription"
)

func (q *Query) CalendarSubscription(ctx context.Context, id string) (*calendarsubscription.CalendarSubscription, error) {
	return q.CalendarSubscriptionStore.FindOne(ctx, id)
}
