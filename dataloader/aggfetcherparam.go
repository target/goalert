package dataloader

import (
	"context"
	"time"
)

// AggFetchParamFunc defines the signature for functions that fetch data with parameters and an array of results.
// The function receives a context, parameter value, and slice of IDs to fetch.
type AggFetchParamFunc[K, P comparable, V any] func(context.Context, P, []K) ([]V, error)

type AggFetchResult[K comparable, V any] struct {
	ID     K
	Values []V
}

type AggFetcherParam[K, P comparable, V any] struct {
	*FetcherParam[K, P, AggFetchResult[K, V]] // Fetcher for aggregated results
}

func NewStoreLoaderAggParam[V any, K, P comparable](ctx context.Context, fetchMany AggFetchParamFunc[K, P, V], idFunc IDFunc[K, V]) *AggFetcherParam[K, P, V] {
	return &AggFetcherParam[K, P, V]{
		FetcherParam: &FetcherParam[K, P, AggFetchResult[K, V]]{
			FetchFunc: func(ctx context.Context, p P, k []K) ([]AggFetchResult[K, V], error) {
				values, err := fetchMany(ctx, p, k)
				if err != nil {
					return nil, err
				}

				result := make(map[K]*AggFetchResult[K, V])
				for _, v := range values {
					id := idFunc(v)
					res, ok := result[id]
					if !ok {
						res = &AggFetchResult[K, V]{ID: id, Values: []V{v}}
						result[id] = res
					} else {
						res.Values = append(res.Values, v)
					}
				}

				resList := make([]AggFetchResult[K, V], 0, len(result))
				for _, v := range result {
					resList = append(resList, *v)
				}
				return resList, nil
			},
			IDFunc:   func(afr AggFetchResult[K, V]) K { return afr.ID },
			MaxBatch: 100,
			Delay:    5 * time.Millisecond,
		},
	}
}

// FetchOneAggParam retrieves a single aggregated value by its ID using the specified parameters.
// Requests with the same parameter values will be batched together, while requests
// with different parameters will use separate batches.
//
// The method lazily creates new Fetcher instances for each unique parameter combination.
func (af *AggFetcherParam[K, P, V]) FetchAggregateParam(ctx context.Context, id K, param P) ([]V, error) {
	res, err := af.FetcherParam.FetchOneParam(ctx, id, param)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil // No results found for this ID
	}
	return res.Values, nil
}
