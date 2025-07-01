package dataloader

import (
	"context"
	"reflect"
	"slices"
	"sync"
	"time"

	"github.com/target/goalert/gadb"
	"github.com/target/goalert/util"
)

func NewStoreLoader[V any, K comparable](ctx context.Context, fetchMany func(context.Context, []K) ([]V, error)) *Fetcher[K, struct{}, V] {
	return &Fetcher[K, struct{}, V]{
		MaxBatch:  100,
		Delay:     time.Millisecond,
		FetchFunc: func(ctx context.Context, s struct{}, k []K) ([]V, error) { return fetchMany(ctx, k) },
	}
}

func NewStoreLoaderWithDB[V any, K comparable](
	ctx context.Context,
	db gadb.DBTX,
	fetchMany func(context.Context, gadb.DBTX, []K) ([]V, error),
) *Fetcher[K, struct{}, V] {
	return NewStoreLoader(ctx, func(ctx context.Context, ids []K) ([]V, error) {
		return fetchMany(ctx, db, ids)
	})
}

type Fetcher[K, P comparable, V any] struct {
	FetchFunc func(ctx context.Context, param P, id []K) ([]V, error)

	IDField string    // If unset defaults to ID, no effect if IDFunc is set.
	IDFunc  func(V) K // Should return the unique ID for a given resource.

	MaxBatch int
	Delay    time.Duration

	cache   map[cacheKey[K, P]]*result[V]
	batches map[P]*batch[K, P, V]
	mx      sync.Mutex
	doInit  sync.Once
	wg      sync.WaitGroup
}

type batch[K, P comparable, V any] struct {
	IDs   []K
	Param P
}
type result[V any] struct {
	value *V
	err   error
	done  chan struct{}
}
type cacheKey[K, P comparable] struct {
	ID    K
	Param P
}

func (f *Fetcher[K, P, V]) init() {
	f.doInit.Do(func() {
		f.cache = make(map[cacheKey[K, P]]*result[V])
		f.batches = make(map[P]*batch[K, P, V])
		if f.IDField == "" {
			f.IDField = "ID"
		}
		if f.IDFunc == nil {
			f.IDFunc = func(v V) K { return reflect.ValueOf(v).FieldByName(f.IDField).Interface().(K) }
		}
	})
}

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
		res, ok := f.cache[cacheKey[K, P]{ID: f.IDFunc(v), Param: param}]
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

func (f *Fetcher[K, P, V]) FetchOne(ctx context.Context, id K) (*V, error) {
	var empty P
	return f.FetchOneParam(ctx, id, empty)
}
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
