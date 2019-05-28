package dataloader

import (
	"context"
	"github.com/target/goalert/user/contactmethod"
	"time"
)

// CMLoader will load user contact methods from postgres.
type CMLoader struct {
	*loader
	store contactmethod.Store
}

// NewCMLoader will create a new CMLoader using the provided store for fetch operations.
func NewCMLoader(ctx context.Context, store contactmethod.Store) *CMLoader {
	p := &CMLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*contactmethod.ContactMethod).ID },
		FetchFunc: p.fetch,
	})
	return p
}

// FetchOne will fetch a single record from the store, batching requests to the store.
func (l *CMLoader) FetchOne(ctx context.Context, id string) (*contactmethod.ContactMethod, error) {
	v, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*contactmethod.ContactMethod), nil
}

func (l *CMLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
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
