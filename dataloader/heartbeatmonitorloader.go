package dataloader

import (
	"context"
	"time"

	"github.com/target/goalert/heartbeat"
)

type HeartbeatMonitorLoader struct {
	*loader
	store heartbeat.Store
}

func NewHeartbeatMonitorLoader(ctx context.Context, store heartbeat.Store) *HeartbeatMonitorLoader {
	p := &HeartbeatMonitorLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*heartbeat.Monitor).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *HeartbeatMonitorLoader) FetchOne(ctx context.Context, id string) (*heartbeat.Monitor, error) {
	v, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*heartbeat.Monitor), nil
}

func (l *HeartbeatMonitorLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
	many, err := l.store.FindMany(ctx, ids...)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(many))
	for i := range many {
		res[i] = &many[i]
	}
	return res, nil
}
