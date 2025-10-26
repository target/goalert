package dataloader

import (
	"context"
	"time"
)

// AggFetcher provides batched loading of aggregated data. Unlike the basic Fetcher
// which returns at most one value per ID, AggFetcher collects multiple values
// for each ID and returns them as a slice.
//
// This is useful for one-to-many relationships where you need to fetch all related
// items for a set of parent IDs, such as fetching all alerts for multiple services.
//
// Type parameters:
//   - K: The type of the unique identifier (key) for items being fetched
//   - V: The type of values being aggregated
type AggFetcher[K comparable, V any] struct {
	*Fetcher[K, AggFetchResult[K, V]] // Embedded Fetcher for aggregated results
}

// AggFetchFunc defines the signature for functions that fetch aggregated data.
// The function receives a context and slice of IDs to fetch, and should return
// all values associated with any of those IDs (potentially multiple values per ID).
type AggFetchFunc[K comparable, V any] func(context.Context, []K) ([]V, error)

// AggFetchResult holds the aggregated results for a single ID.
// It contains the ID and all values associated with that ID.
type AggFetchResult[K comparable, V any] struct {
	ID     K   // The unique identifier
	Values []V // All values associated with this ID
}

// NewStoreLoaderAgg creates a new AggFetcher for loading aggregated data from a store.
// This is useful for one-to-many relationships where you need to batch-load multiple
// related items for each parent ID.
//
// The fetchMany function should return all values for the given IDs. Values can be
// returned in any order and multiple values can have the same ID (as determined by idFunc).
// The AggFetcher will automatically group them by ID.
//
// Example usage:
//
//	// Fetch all alerts for multiple services
//	alertLoader := dataloader.NewStoreLoaderAgg(ctx,
//		func(ctx context.Context, serviceIDs []string) ([]Alert, error) {
//			return alertStore.FindByServiceIDs(ctx, serviceIDs)
//		},
//		func(alert Alert) string { return alert.ServiceID },
//	)
//	alerts, err := alertLoader.FetchAggregate(ctx, "service-123")
func NewStoreLoaderAgg[V any, K comparable](ctx context.Context, fetchMany AggFetchFunc[K, V], idFunc IDFunc[K, V]) *AggFetcher[K, V] {
	return &AggFetcher[K, V]{
		Fetcher: &Fetcher[K, AggFetchResult[K, V]]{
			FetchFunc: func(ctx context.Context, ids []K) ([]AggFetchResult[K, V], error) {
				// Fetch all values for the requested IDs
				values, err := fetchMany(ctx, ids)
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

// FetchAggregate retrieves all values associated with the given ID.
// Multiple calls within the same batch window will be batched together
// into a single call to the underlying fetch function.
//
// Returns a slice of all values associated with the ID, or nil if no values
// are found. An error is returned if the fetch operation fails or the context
// is cancelled.
func (af *AggFetcher[K, V]) FetchAggregate(ctx context.Context, id K) ([]V, error) {
	res, err := af.FetchOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, nil // No results found for this ID
	}
	return res.Values, nil
}
