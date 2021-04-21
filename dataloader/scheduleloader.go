package dataloader

import (
	"context"
	"time"

	"github.com/target/goalert/schedule"
)

type ScheduleLoader struct {
	*loader
	store *schedule.Store
}

func NewScheduleLoader(ctx context.Context, store *schedule.Store) *ScheduleLoader {
	p := &ScheduleLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*schedule.Schedule).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *ScheduleLoader) FetchOne(ctx context.Context, id string) (*schedule.Schedule, error) {
	v, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*schedule.Schedule), nil
}

func (l *ScheduleLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
	many, err := l.store.FindMany(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(many))
	for i := range many {
		res[i] = &many[i]
	}
	return res, nil
}
