package dataloader

import (
	"context"
	"time"
)

type AggFetchResult[K comparable, V any] struct {
	ID     K
	Values []V
}

type AggFetcher[K, P comparable, V any] struct {
	f       *Fetcher[K, P, AggFetchResult[K, V]]
	IDField string
	IDFunc  func(V) K
}

// SetIDField sets the name of the field to use as the unique identifier for values.
// This method must be called before any FetchOne calls. The field must be exported
// and accessible via reflection.
//
// Returns the Fetcher instance for method chaining.
func (a *AggFetcher[K, P, V]) SetIDField(field string) *AggFetcher[K, P, V] {
	a.IDField = field
	return a
}

// SetIDFunc sets the function to extract the unique identifier from a value.
// When set, this takes precedence over IDField and is more efficient than reflection.
// This method must be called before any FetchOne calls.
//
// Returns the Fetcher instance for method chaining.
func (a *AggFetcher[K, P, V]) SetIDFunc(fn func(V) K) *AggFetcher[K, P, V] {
	a.IDFunc = fn
	return a
}

// LookupID extracts the unique identifier from a value using the configured IDFunc or IDField.
func (f *AggFetcher[K, P, V]) LookupID(v V) K { return LookupID(v, f.IDField, f.IDFunc) }

func NewStoreLoaderAggParam[V any, K, P comparable](ctx context.Context, fetchMany func(context.Context, P, []K) ([]V, error)) *AggFetcher[K, P, V] {
	a := &AggFetcher[K, P, V]{
		f: &Fetcher[K, P, AggFetchResult[K, V]]{
			MaxBatch: 100,
			Delay:    time.Millisecond,
			IDFunc:   func(afr AggFetchResult[K, V]) K { return afr.ID },
		},
	}

	a.f.FetchFunc = func(ctx context.Context, params P, ids []K) ([]AggFetchResult[K, V], error) {
		values, err := fetchMany(ctx, params, ids)
		if err != nil {
			return nil, err
		}
		results := make(map[K][]V)
		for _, v := range values {
			id := a.LookupID(v)
			results[id] = append(results[id], v)
		}
		res := make([]AggFetchResult[K, V], 0, len(ids))
		for id, vals := range results {
			res = append(res, AggFetchResult[K, V]{ID: id, Values: vals})
		}
		return res, nil
	}

	return a
}

func NewStoreLoaderAgg[V any, K comparable](ctx context.Context, fetchMany func(context.Context, []K) ([]V, error)) *AggFetcher[K, struct{}, V] {
	return NewStoreLoaderAggParam(ctx, func(ctx context.Context, _ struct{}, ids []K) ([]V, error) {
		return fetchMany(ctx, ids)
	})
}

func (a *AggFetcher[K, P, V]) FetchOne(ctx context.Context, id K) ([]V, error) {
	res, err := a.f.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	return res.Values, nil
}

func (a *AggFetcher[K, P, V]) FetchOneParam(ctx context.Context, id K, param P) ([]V, error) {
	res, err := a.f.FetchOneParam(ctx, id, param)
	if err != nil {
		return nil, err
	}
	return res.Values, nil
}
