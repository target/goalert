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
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util"
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
func NewStoreLoader[V any, K comparable](ctx context.Context, fetchMany func(context.Context, []K) ([]V, error)) *Fetcher[K, struct{}, V] {
	return &Fetcher[K, struct{}, V]{
		MaxBatch:  100,
		Delay:     time.Millisecond,
		FetchFunc: func(ctx context.Context, s struct{}, ids []K) ([]V, error) { return fetchMany(ctx, ids) },
	}
}

// NewStoreLoaderWithDB creates a new Fetcher that automatically passes a database
// connection to the fetch function. This is a convenience wrapper around NewStoreLoader
// for database-based loading operations.
func NewStoreLoaderWithDB[V any, K comparable](
	ctx context.Context,
	db gadb.DBTX,
	fetchMany func(context.Context, gadb.DBTX, []K) ([]V, error),
) *Fetcher[K, struct{}, V] {
	return NewStoreLoader(ctx, func(ctx context.Context, ids []K) ([]V, error) {
		return fetchMany(ctx, db, ids)
	})
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
type Fetcher[K, P comparable, V any] struct {
	// FetchFunc is called to retrieve data for a batch of IDs with the given parameters.
	// It should return values in any order - the Fetcher will map them back to requests
	// using the ID extracted via IDFunc or IDField.
	FetchFunc func(ctx context.Context, param P, ids []K) ([]V, error)

	// IDField specifies the name of the field to use as the unique identifier.
	// Defaults to "ID" if not set and IDFunc is nil. This field is accessed via
	// reflection, so it must be exported.
	IDField string

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

	cache   map[cacheKey[K, P]]*result[V]
	batches map[P]*batch[K, P, V]
	mx      sync.Mutex
	doInit  sync.Once
	wg      sync.WaitGroup
}

// batch represents a group of IDs that should be fetched together with the same parameters.
type batch[K, P comparable, V any] struct {
	IDs   []K
	Param P
}

// result holds the outcome of fetching a single item, including the value, any error,
// and a channel to signal completion.
type result[V any] struct {
	value *V
	err   error
	done  chan struct{}
}

// cacheKey uniquely identifies a cached item by its ID and parameters.
type cacheKey[K, P comparable] struct {
	ID    K
	Param P
}

// DefaultIDFunc provides a default implementation for extracting the ID from a struct.
func DefaultIDFunc[K comparable, V any](v V, idField string) (id K) {
	val := reflect.ValueOf(v).FieldByName(idField)
	if !val.IsValid() || !val.CanInterface() {
		return id // empty/zero value if field is not accessible
	}

	if v, ok := val.Interface().(K); ok {
		return v // directly return if type matches, this is the common case
	}

	if reflect.TypeOf(id) == reflect.TypeOf(int(0)) {
		// common case when K is int and val is int or int64
		if val.Kind() == reflect.Int || val.Kind() == reflect.Int64 {
			return any(val.Int()).(K)
		}
	}

	// special case when K is string and val is uuid.UUID
	if val.Type() == reflect.TypeOf(uuid.UUID{}) && reflect.TypeOf(id) == reflect.TypeOf("") {
		s := val.Interface().(uuid.UUID).String()
		return any(s).(K)
	}

	// inverse case when K is uuid.UUID and val is string
	if val.Type() == reflect.TypeOf("") && reflect.TypeOf(id) == reflect.TypeOf(uuid.UUID{}) {
		uuidStr, ok := val.Interface().(string)
		if !ok {
			return id // empty/zero value if conversion fails
		}
		uuidVal, err := uuid.Parse(uuidStr)
		if err != nil {
			return id // empty/zero value if conversion fails
		}
		return any(uuidVal).(K)
	}

	// fallback to using the zero value of K if type conversion fails
	return id
}

func (f *Fetcher[K, P, V]) init() {
	f.doInit.Do(func() {
		f.cache = make(map[cacheKey[K, P]]*result[V])
		f.batches = make(map[P]*batch[K, P, V])
	})
}

func LookupID[K comparable, V any](v V, idField string, idFunc func(V) K) K {
	if idFunc != nil {
		return idFunc(v)
	}
	if idField == "" {
		idField = "ID" // default field name if not set
	}

	if ider, ok := any(v).(IDer[K]); ok {
		return ider.ID() // use IDer interface if available
	}

	return DefaultIDFunc[K](v, idField)
}

// LookupID extracts the unique identifier from a value using the configured IDFunc or IDField.
func (f *Fetcher[K, P, V]) LookupID(v V) K { return LookupID(v, f.IDField, f.IDFunc) }

// Close waits for all pending batches to complete. This should be called when
// the Fetcher is no longer needed to ensure proper cleanup and prevent
// goroutine leaks.
func (f *Fetcher[K, P, V]) Close() {
	// Wait for all batches to complete
	f.wg.Wait()
}

func (f *Fetcher[K, P, V]) _batch(ctx context.Context, param P, id K) {
	b, ok := f.batches[param]
	if !ok || len(b.IDs) >= f.MaxBatch {
		b = &batch[K, P, V]{Param: param, IDs: []K{id}}
		f.batches[param] = b
		f.wg.Add(1)
		go f.runBatch(ctx, param, b)
	} else if !slices.Contains(b.IDs, id) {
		b.IDs = append(b.IDs, id)
	}
}

func (f *Fetcher[K, P, V]) runBatch(ctx context.Context, param P, b *batch[K, P, V]) {
	defer f.wg.Done()
	_ = util.ContextSleep(ctx, f.Delay)

	f.mx.Lock()
	delete(f.batches, param)
	f.mx.Unlock()

	var values []V
	values, err := f.FetchFunc(ctx, param, b.IDs)

	f.mx.Lock()
	for _, v := range values {
		res, ok := f.cache[cacheKey[K, P]{ID: f.LookupID(v), Param: param}]
		if !ok {
			// we didn't ask for this ID, ignore it
			continue
		}
		if res.done == nil {
			// just in case there was a duplicate somehow
			continue
		}

		if err != nil {
			res.err = err
		} else {
			res.value = &v
		}
		close(res.done)
		res.done = nil
	}
	// remaining were not found, mark as done
	for _, id := range b.IDs {
		res := f.cache[cacheKey[K, P]{ID: id, Param: param}]
		if res.done == nil {
			// just in case there was a duplicate somehow
			continue
		}

		if err != nil {
			res.err = err
		}
		close(res.done)
		res.done = nil
	}
	f.mx.Unlock()
}

// FetchOne retrieves a single value by its ID using default (empty) parameters.
// This is a convenience method equivalent to calling FetchOneParam with an empty parameter.
//
// The method returns a pointer to the value if found, or nil if not found.
// An error is returned if the fetch operation fails or the context is cancelled.
func (f *Fetcher[K, P, V]) FetchOne(ctx context.Context, id K) (*V, error) {
	var empty P
	return f.FetchOneParam(ctx, id, empty)
}

// FetchOneParam retrieves a single value by its ID with additional parameters.
// The request may be batched with other concurrent requests that have the same parameters.
//
// Parameters with different values will result in separate batches, allowing for
// parameter-specific optimizations in the fetch function.
//
// The method returns a pointer to the value if found, or nil if not found.
// An error is returned if the fetch operation fails or the context is cancelled.
func (f *Fetcher[K, P, V]) FetchOneParam(ctx context.Context, id K, param P) (*V, error) {
	f.init()

	f.mx.Lock()
	r, ok := f.cache[cacheKey[K, P]{ID: id, Param: param}]
	if !ok {
		r = &result[V]{done: make(chan struct{})}
		f.cache[cacheKey[K, P]{ID: id, Param: param}] = r
		f._batch(ctx, param, id)
	}

	f.mx.Unlock()
	select {
	case <-r.done:
		return r.value, r.err
	case <-ctx.Done():
		return r.value, ctx.Err()
	}
}
