package dataloader

import (
	"context"
	"github.com/target/goalert/user"
	"time"
)

type UserLoader struct {
	*loader
	store user.Store
}

func NewUserLoader(ctx context.Context, store user.Store) *UserLoader {
	p := &UserLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*user.User).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (l *UserLoader) FetchOne(ctx context.Context, id string) (*user.User, error) {
	v, err := l.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, err
	}
	return v.(*user.User), nil
}

func (l *UserLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
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
