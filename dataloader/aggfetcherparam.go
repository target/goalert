package dataloader

import (
	"context"
	"time"
)

// AggFetchParamFunc defines the signature for functions that fetch aggregated data with parameters.
// The function receives a context, parameter value, and slice of IDs to fetch, and should return
// all values associated with any of those IDs for the given parameters.
type AggFetchParamFunc[K, P comparable, V any] func(context.Context, P, []K) ([]V, error)

// AggFetcherParam provides batched loading of aggregated data with parameters.
// Unlike AggFetcher, this allows you to pass additional parameters that modify
// the fetch behavior. It maintains separate Fetcher instances for each unique
// parameter combination.
//
// This is useful for parameterized one-to-many relationships, such as fetching
// all active alerts for multiple services, or all escalation policies for
// multiple teams with specific filters.
//
// Type parameters:
//   - K: The type of the unique identifier (key) for items being fetched
//   - P: The type of additional parameters that can be passed to modify fetch behavior
//   - V: The type of values being aggregated
type AggFetcherParam[K, P comparable, V any] struct {
	*FetcherParam[K, P, AggFetchResult[K, V]] // Embedded FetcherParam for aggregated results
}

// NewStoreLoaderAggParam creates a new AggFetcherParam for loading aggregated data
// from a store with parameters. This allows you to batch requests not just by ID,
// but also by parameter values, while still aggregating multiple results per ID.
//
// The fetchMany function should handle the parameter and return all values for
// the given IDs that match the parameter criteria. Values can be returned in any
// order and multiple values can have the same ID (as determined by idFunc).
// The AggFetcherParam will automatically group them by ID.
//
// Example usage:
//
//	type AlertParams struct { Status string }
//	alertLoader := dataloader.NewStoreLoaderAggParam(ctx,
//		func(ctx context.Context, params AlertParams, serviceIDs []string) ([]Alert, error) {
//			return alertStore.FindByServiceIDsAndStatus(ctx, serviceIDs, params.Status)
//		},
//		func(alert Alert) string { return alert.ServiceID },
//	)
//	openAlerts, err := alertLoader.FetchAggregateParam(ctx, "service-123", AlertParams{Status: "open"})
func NewStoreLoaderAggParam[V any, K, P comparable](ctx context.Context, fetchMany AggFetchParamFunc[K, P, V], idFunc IDFunc[K, V]) *AggFetcherParam[K, P, V] {
	return &AggFetcherParam[K, P, V]{
		FetcherParam: &FetcherParam[K, P, AggFetchResult[K, V]]{
			FetchFunc: func(ctx context.Context, p P, k []K) ([]AggFetchResult[K, V], error) {
				// Fetch all values for the requested IDs with the given parameters
				values, err := fetchMany(ctx, p, k)
				if err != nil {
					return nil, err
				}

				// Group values by their ID
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

				// Convert map to slice for return
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

// FetchAggregateParam retrieves all values associated with the given ID using the specified parameters.
// Requests with the same parameter values will be batched together, while requests
// with different parameters will use separate batches.
//
// The method lazily creates new Fetcher instances for each unique parameter combination.
// Returns a slice of all values associated with the ID for the given parameters,
// or nil if no values are found. An error is returned if the fetch operation fails
// or the context is cancelled.
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
