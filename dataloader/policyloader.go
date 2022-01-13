package dataloader

import (
	"context"
	"github.com/target/goalert/escalation"
	"time"
)

type PolicyLoader struct {
	*loader
	store escalation.Store
}

func NewPolicyLoader(ctx context.Context, store escalation.Store) *PolicyLoader {
	p := &PolicyLoader{
		store: store,
	}
	p.loader = newLoader(ctx, loaderConfig{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(pol interface{}) string { return pol.(*escalation.Policy).ID },
		FetchFunc: p.fetch,
	})
	return p
}

func (p *PolicyLoader) FetchOne(ctx context.Context, id string) (*escalation.Policy, error) {
	pol, err := p.loader.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if pol == nil {
		return nil, err
	}
	return pol.(*escalation.Policy), nil
}

func (p *PolicyLoader) fetch(ctx context.Context, ids []string) ([]interface{}, error) {
	pol, err := p.store.FindManyPolicies(ctx, ids)
	if err != nil {
		return nil, err
	}

	res := make([]interface{}, len(pol))
	for i := range pol {
		res[i] = &pol[i]
	}
	return res, nil
}
