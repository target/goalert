package dataloader

import (
	"context"
	"time"

	"github.com/target/goalert/notification"
)

type AlertLogMessageStatusLoader struct {
	*loader
	store notification.Store
}

func NewAlertLogMessageStatusLoader(ctx context.Context, store notification.Store) *AlertLogMessageStatusLoader {
	p := &AlertLogMessageStatusLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*notification.MessageStatus).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *AlertLogMessageStatusLoader) FetchOne(ctx context.Context, id string) (*notification.MessageStatus, error) {
	ls, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if ls == nil {
		return nil, err
	}
	return ls.(*notification.MessageStatus), nil
}

func (l *AlertLogMessageStatusLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
	many, err := l.store.FindManyMessageStatuses(ctx, ids...)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(many))
	for i := range many {
		res[i] = &many[i]
	}
	return res, nil
}
