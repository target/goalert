package dataloader

import (
	"context"
	"sync"
	"time"
)

// FetcherParam provides batched loading of data with parameters. Unlike the basic Fetcher,
// this allows you to pass additional parameters that modify the fetch behavior.
// It maintains separate Fetcher instances for each unique parameter combination.
//
// Type parameters:
//   - K: The type of the unique identifier (key) for items being fetched
//   - P: The type of additional parameters that can be passed to modify fetch behavior
//   - V: The type of values being fetched
//
// Example usage:
//
//	type UserParams struct { Active bool }
//	paramLoader := dataloader.NewStoreLoaderParam(
//	  ctx,
//	  func(ctx context.Context, param UserParams, ids []string) ([]User, error) {
//	    return userStore.FindManyFiltered(ctx, ids, param.Active)
//	  },
//	  func(u User) string { return u.ID },
//	)
//	activeUser, err := paramLoader.FetchOneParam(ctx, "user-id", UserParams{Active: true})
type FetcherParam[K, P comparable, V any] struct {
	mx        sync.Mutex              // Protects the map of fetchers
	m         map[P]*Fetcher[K, V]    // Map of parameters to their corresponding fetchers
	FetchFunc FetchParamFunc[K, P, V] // Function called to fetch data with parameters
	IDFunc    IDFunc[K, V]            // Function to extract ID from values
	MaxBatch  int                     // Maximum batch size for each parameter's fetcher
	Delay     time.Duration           // Delay before executing batches
}

// FetchParamFunc defines the signature for functions that fetch data with parameters.
// The function receives a context, parameter value, and slice of IDs to fetch.
type FetchParamFunc[K, P comparable, V any] func(context.Context, P, []K) ([]V, error)

// NewStoreLoaderParam creates a new FetcherParam for loading data from a store with parameters.
// This allows you to batch requests not just by ID, but also by parameter values.
//
// The fetchMany function should handle the parameter and return values in any order.
// The FetcherParam will create separate Fetcher instances for each unique parameter combination.
func NewStoreLoaderParam[K, P comparable, V any](ctx context.Context, fetchMany FetchParamFunc[K, P, V], idFunc IDFunc[K, V]) *FetcherParam[K, P, V] {
	return &FetcherParam[K, P, V]{
		m:         make(map[P]*Fetcher[K, V]),
		FetchFunc: fetchMany,
		IDFunc:    idFunc,
		MaxBatch:  100,
		Delay:     5 * time.Millisecond,
	}
}

// FetchOneParam retrieves a single value by its ID using the specified parameters.
// Requests with the same parameter values will be batched together, while requests
// with different parameters will use separate batches.
//
// The method lazily creates new Fetcher instances for each unique parameter combination.
// This allows efficient batching while still supporting parameterized queries.
func (fp *FetcherParam[K, P, V]) FetchOneParam(ctx context.Context, id K, param P) (*V, error) {
	fp.mx.Lock()
	defer fp.mx.Unlock()

	// Get or create a Fetcher for this parameter combination
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

func (fp *FetcherParam[K, P, V]) Close() {
	fp.mx.Lock()
	defer fp.mx.Unlock()

	// Close all fetchers for each parameter
	for _, f := range fp.m {
		f.Close()
	}
}
