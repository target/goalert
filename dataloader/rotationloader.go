package dataloader

import (
	"context"
	"github.com/target/goalert/schedule/rotation"
	"time"
)

type RotationLoader struct {
	*loader
	store rotation.Store
}

func NewRotationLoader(ctx context.Context, store rotation.Store) *RotationLoader {
	p := &RotationLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*rotation.Rotation).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *RotationLoader) FetchOne(ctx context.Context, id string) (*rotation.Rotation, error) {
	v, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*rotation.Rotation), nil
}

func (l *RotationLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
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
