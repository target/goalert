// Package dataloader provides a generic batching and caching mechanism for loading data
// efficiently. It reduces the number of database queries by batching multiple individual
// requests together and caching results to avoid duplicate fetches.
//
// The primary use case is in GraphQL resolvers where you might have N+1 query problems.
// Instead of making individual database calls for each item, the dataloader batches
// requests together and executes them in a single operation.
//
// Example usage:
//
//	// Create a loader for User entities
//	userLoader := dataloader.NewStoreLoader(ctx, userStore.FindMany)
//	userLoader.MaxBatch = 100
//	userLoader.Delay = 5 * time.Millisecond
//
//	// Use in resolvers - these calls will be batched together
//	user1, err := userLoader.FetchOne(ctx, "user-id-1")
//	user2, err := userLoader.FetchOne(ctx, "user-id-2")
//	user3, err := userLoader.FetchOne(ctx, "user-id-3")
//
//	// With parameters for filtered loading
//	type UserParams struct { Active bool }
//	paramLoader := &Fetcher[string, UserParams, User]{
//		FetchFunc: func(ctx context.Context, param UserParams, ids []string) ([]User, error) {
//			return userStore.FindManyFiltered(ctx, ids, param.Active)
//		},
//		MaxBatch: 50,
//		Delay: time.Millisecond,
//	}
//	activeUser, err := paramLoader.FetchOneParam(ctx, "user-id", UserParams{Active: true})
package dataloader

import (
	"context"
	"sync"
	"time"

	"github.com/target/goalert/gadb"
)

type Loader[K comparable, V any] = Fetcher[K, V]

type (
	FetchFunc[K comparable, V any] func(context.Context, []K) ([]V, error)
	IDFunc[K comparable, V any]    func(V) K
)

type IDer[K comparable] interface {
	// ID returns the unique identifier for the value.
	ID() K
}

// NewStoreLoader creates a new Fetcher for loading data from a store without parameters.
// It's a convenience function for the common case where you only need to batch by ID.
//
// The fetchMany function should return values in any order, and the Fetcher will
// map them back to the correct requests using the ID field.
func NewStoreLoader[V any, K comparable](ctx context.Context, fetchMany FetchFunc[K, V], idFunc IDFunc[K, V]) *Fetcher[K, V] {
	return &Fetcher[K, V]{
		MaxBatch:  100,
		Delay:     5 * time.Millisecond,
		FetchFunc: func(ctx context.Context, ids []K) ([]V, error) { return fetchMany(ctx, ids) },
		IDFunc:    idFunc,
	}
}

// NewStoreLoaderWithDB creates a new Fetcher that automatically passes a database
// connection to the fetch function. This is a convenience wrapper around NewStoreLoader
// for database-based loading operations.
func NewStoreLoaderWithDB[V any, K comparable](
	ctx context.Context,
	db gadb.DBTX,
	fetchMany func(context.Context, gadb.DBTX, []K) ([]V, error),
	idFunc IDFunc[K, V],
) *Fetcher[K, V] {
	return NewStoreLoader(ctx, func(ctx context.Context, ids []K) ([]V, error) {
		return fetchMany(ctx, db, ids)
	}, idFunc)
}

// Fetcher provides batched loading of data with caching. It batches individual requests
// together to reduce the number of calls to the underlying data source, and caches
// results to avoid duplicate fetches within the same Fetcher instance.
//
// Type parameters:
//   - K: The type of the unique identifier (key) for items being fetched
//   - P: The type of additional parameters that can be passed to modify fetch behavior
//   - V: The type of values being fetched
//
// The Fetcher is safe for concurrent use. All methods can be called from multiple
// goroutines simultaneously.
type Fetcher[K comparable, V any] struct {
	// FetchFunc is called to retrieve data for a batch of IDs with the given parameters.
	// It should return values in any order - the Fetcher will map them back to requests
	// using the ID extracted via IDFunc or IDField.
	FetchFunc func(ctx context.Context, ids []K) ([]V, error)

	// IDFunc extracts the unique identifier from a value. If set, this takes
	// precedence over IDField. This is more efficient than using reflection
	// and allows for more complex ID extraction logic.
	IDFunc func(V) K

	// MaxBatch sets the maximum number of IDs to include in a single batch.
	// When this limit is reached, a new batch is started immediately.
	MaxBatch int

	// Delay specifies how long to wait before executing a batch. This allows
	// multiple requests to accumulate into a single batch, improving efficiency.
	Delay time.Duration

	cache map[K]*result[K, V]

	currentBatch *batch[K, V]
	mx           sync.Mutex
	wg           sync.WaitGroup
}

// result holds the outcome of fetching a single item, including the value, any error,
// and a channel to signal completion.
type result[K comparable, V any] struct {
	id    K
	value *V
	err   error
	done  chan struct{}
}

func (r *result[K, V]) wait(ctx context.Context) (*V, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-r.done:
		return r.value, r.err
	}
}

// Close waits for all pending batches to complete. This should be called when
// the Fetcher is no longer needed to ensure proper cleanup and prevent
// goroutine leaks.
func (f *Fetcher[K, V]) Close() {
	// Wait for all batches to complete
	f.wg.Wait()
}

func (f *Fetcher[K, V]) fetchAll(ctx context.Context, batch *batch[K, V]) {
	defer f.wg.Done()

	// Since we are exlusive to this batch, we fetch before locking the cache.
	values, err := f.FetchFunc(ctx, batch.ids)
	f.mx.Lock()
	defer f.mx.Unlock()
	if err != nil {
		// In the error case, we close all results with the error since there is no value to return.
		for _, res := range batch.results {
			res.err = err
			close(res.done)
		}
		return
	}

	for _, val := range values {
		id := f.IDFunc(val)
		res, ok := f.cache[id]
		if !ok {
			// we didn't ask for this ID, ignore it
			continue
		}

		res.value = &val
	}

	// We close all results in a separate loop for cases where a value is missing. Since the batch is done, all results must be closed.
	for _, res := range batch.results {
		close(res.done)
	}
}

// FetchOne retrieves a single value by its ID using default (empty) parameters.
// This is a convenience method equivalent to calling FetchOneParam with an empty parameter.
//
// The method returns a pointer to the value if found, or nil if not found.
// An error is returned if the fetch operation fails or the context is cancelled.
func (f *Fetcher[K, V]) FetchOne(ctx context.Context, id K) (*V, error) {
	f.mx.Lock()

	if f.cache == nil {
		f.cache = make(map[K]*result[K, V])
	}

	if res, ok := f.cache[id]; ok {
		// easy path, already cached
		f.mx.Unlock()
		return res.wait(ctx)
	}

	// create the entry for the new ID
	if f.currentBatch == nil {
		f.currentBatch = new(batch[K, V])
		f.wg.Add(1)
		time.AfterFunc(f.Delay, func() {
			f.fetchAll(ctx, f.currentBatch)
		})
	}
	res := f.currentBatch.Add(ctx, id)
	f.cache[id] = res

	if len(f.currentBatch.ids) >= f.MaxBatch {
		f.currentBatch = nil // reset for next batch
	}

	f.mx.Unlock()
	return res.wait(ctx)
}
