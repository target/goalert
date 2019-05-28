package dataloader

import (
	"context"
	"github.com/target/goalert/service"
	"time"
)

type ServiceLoader struct {
	*loader
	store service.Store
}

func NewServiceLoader(ctx context.Context, store service.Store) *ServiceLoader {
	p := &ServiceLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*service.Service).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *ServiceLoader) FetchOne(ctx context.Context, id string) (*service.Service, error) {
	svc, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, err
	}
	return svc.(*service.Service), nil
}

func (l *ServiceLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
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
