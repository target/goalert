package dataloader

import (
	"context"
	"sync"
	"time"
)

type FetcherParam[K, P comparable, V any] struct {
	mx        sync.Mutex
	m         map[P]*Fetcher[K, V]
	FetchFunc FetchParamFunc[K, P, V]
	IDFunc    IDFunc[K, V]
	MaxBatch  int
	Delay     time.Duration
}

type FetchParamFunc[K, P comparable, V any] func(context.Context, P, []K) ([]V, error)

// NewStoreLoaderParam creates a new Fetcher for loading data from a store with parameters.
func NewStoreLoaderParam[K, P comparable, V any](ctx context.Context, fetchMany FetchParamFunc[K, P, V], idFunc IDFunc[K, V]) *FetcherParam[K, P, V] {
	return &FetcherParam[K, P, V]{
		m:         make(map[P]*Fetcher[K, V]),
		FetchFunc: fetchMany,
		IDFunc:    idFunc,
		MaxBatch:  100,
		Delay:     5 * time.Millisecond,
	}
}

func (fp *FetcherParam[K, P, V]) FetchOneParam(ctx context.Context, id K, param P) (*V, error) {
	fp.mx.Lock()
	defer fp.mx.Unlock()

	loader, ok := fp.m[param]
	if !ok {
		loader = &Fetcher[K, V]{
			MaxBatch:  fp.MaxBatch,
			Delay:     fp.Delay,
			FetchFunc: func(ctx context.Context, ids []K) ([]V, error) { return fp.FetchFunc(ctx, param, ids) },
			IDFunc:    fp.IDFunc,
		}
		fp.m[param] = loader
	}

	return loader.FetchOne(ctx, id)
}
