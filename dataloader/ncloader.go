package dataloader

import (
	"context"
	"time"

	"github.com/target/goalert/notificationchannel"
)

// NCLoader will load notification channels from postgres.
type NCLoader struct {
	*loader
	store notificationchannel.Store
}

// NewNCLoader will create a new CMLoader using the provided store for fetch operations.
func NewNCLoader(ctx context.Context, store notificationchannel.Store) *NCLoader {
	p := &NCLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*notificationchannel.Channel).ID },
		FetchFunc: p.fetch,
	})
	return p
}

// FetchOne will fetch a single record from the store, batching requests to the store.
func (l *NCLoader) FetchOne(ctx context.Context, id string) (*notificationchannel.Channel, error) {
	v, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*notificationchannel.Channel), nil
}

func (l *NCLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
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
