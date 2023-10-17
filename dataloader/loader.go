package dataloader

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/target/goalert/gadb"
)

func NewStoreLoaderInt[V any](ctx context.Context, fetchMany func(context.Context, []int) ([]V, error)) *Loader[int, V] {
	return newLoader(ctx, loaderConfig[int, V]{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v V) int { return int(reflect.ValueOf(v).FieldByName("ID").Int()) },
		FetchFunc: fetchMany,
	})
}

func NewStoreLoader[V any](ctx context.Context, fetchMany func(context.Context, []string) ([]V, error)) *Loader[string, V] {
	return newLoader(ctx, loaderConfig[string, V]{
		Max:       100,
		Delay:     time.Millisecond,
		IDFunc:    func(v V) string { return reflect.ValueOf(v).FieldByName("ID").String() },
		FetchFunc: fetchMany,
	})
}

func NewStoreLoaderWithDB[V any](
	ctx context.Context,
	db gadb.DBTX,
	fetchMany func(context.Context, gadb.DBTX, []string) ([]V, error),
) *Loader[string, V] {
	return NewStoreLoader(ctx, func(ctx context.Context, ids []string) ([]V, error) {
		return fetchMany(ctx, db, ids)
	})
}

type loaderConfig[K comparable, V any] struct {
	FetchFunc func(context.Context, []K) ([]V, error) // FetchFunc should return resources for the provided IDs (order doesn't matter).
	IDFunc    func(V) K                               // Should return the unique ID for a given resource.

	Delay time.Duration // Delay before fetching pending requests.
	Max   int           // Max number of pending requests before immediate fetch (also max number of requests per-db-call).

	Name string // Name to use in traces.
}

type loaderReq[K comparable, V any] struct {
	id K
	ch chan *entry[K, V]
}

type entry[K comparable, V any] struct {
	id   K
	done chan struct{}
	err  error
	data *V
}
type Loader[K comparable, V any] struct {
	ctx    context.Context
	cancel func()

	cfg   loaderConfig[K, V]
	cache map[K]*entry[K, V]

	start sync.Once

	reqCh chan loaderReq[K, V]
}

func newLoader[K comparable, V any](ctx context.Context, cfg loaderConfig[K, V]) *Loader[K, V] {
	l := &Loader[K, V]{cfg: cfg}
	l.ctx, l.cancel = context.WithCancel(ctx)

	return l
}

func (l *Loader[K, V]) Close() error {
	// ensure we don't start in the future, ensure cancel is called
	// before `start.Do` returns if it's the first call.
	l.start.Do(l.cancel)

	// always call l.cancel
	l.cancel()

	return nil
}

func (l *Loader[K, V]) init() {
	l.cache = make(map[K]*entry[K, V], l.cfg.Max)
	l.reqCh = make(chan loaderReq[K, V], l.cfg.Max)
	go l.loop()
}

// load will perform a batch load for a list of entries
func (l *Loader[K, V]) load(entries []*entry[K, V]) []*entry[K, V] {
	// If we need to load more than the max, call load with the max, and return the rest.
	if len(entries) > l.cfg.Max {
		l.load(entries[:l.cfg.Max])
		return entries[l.cfg.Max:]
	}

	// We need to copy the list so we don't get overwritten if other
	// batch updates are done in the background while this call is processing.
	cpy := make([]*entry[K, V], len(entries))
	copy(cpy, entries)

	go func() {
		ctx := l.ctx

		// Map the entries out by ID, and collect the list of IDs
		// for the DB call.
		m := make(map[K]*entry[K, V], len(entries))
		ids := make([]K, len(entries))
		for i, e := range cpy {
			ids[i] = e.id
			m[e.id] = e
		}

		// Call fetch for everything we're loading
		res, err := l.cfg.FetchFunc(ctx, ids)
		if err != nil {
			// If the fetch failed, set all the pending entries err property to
			// reflect the failure, and close the done channel to indicate the load/fetch
			// completed.
			for _, e := range cpy {
				e.err = err
				close(e.done)
			}
			return
		}

		for i := range res {
			// Go through each received response and update the data value based on the ID.
			// We're processing against a map so we can ignore order within fetch methods.
			id := l.cfg.IDFunc(res[i])
			e := m[id]
			if e == nil {
				// Ignore any unknown/unexpected results
				continue
			}
			e.data = &res[i]
		}

		// nil or not, all entries are now done loading, if the .data prop was not set
		// then the entry does not exist.
		for _, e := range cpy {
			close(e.done)
		}
	}()

	return entries[:0]
}

// entry will return the current entry or create a new one in the map.
// It passes the new or existing loaderEntry to the requester.
func (l *Loader[K, V]) entry(req loaderReq[K, V]) (*entry[K, V], bool) {
	if v, ok := l.cache[req.id]; ok {
		req.ch <- v
		return v, false
	}

	e := &entry[K, V]{
		id:   req.id,
		done: make(chan struct{}),
	}
	l.cache[req.id] = e
	req.ch <- e
	return e, true
}

func (l *Loader[K, V]) loop() {
	// timerStart tracks if the delay timer has started or not.
	var timerStart bool
	var t *time.Timer
	// waitCh by default will block indefinitely, since the timer
	// shouldn't start until the first pending request is made.
	waitCh := (<-chan time.Time)(make(chan time.Time))
	batch := make([]*entry[K, V], 0, l.cfg.Max)

	var req loaderReq[K, V]
	for {
		select {
		case <-waitCh:
			// timer expired load immediately
			batch = l.load(batch)
			timerStart = false
		case <-l.ctx.Done():
			return
		case req = <-l.reqCh:
			e, isNew := l.entry(req)
			if !isNew {
				// request for that ID is already pending, nothing to do
				continue
			}

			// new entries get added to the batch
			batch = append(batch, e)
		}

		// If we ever exceed max, immediately load, batch then becomes
		// whatever is left.
		if len(batch) > l.cfg.Max {
			batch = l.load(batch)
		}

		if !timerStart && len(batch) > 0 {
			// If the timer hasn't started, but we have something waiting, start it
			timerStart = true
			t = time.NewTimer(l.cfg.Delay)
			waitCh = t.C
		} else if timerStart && len(batch) == 0 {
			// If the timer is running, but there are no more pending entries,
			// stop it.
			t.Stop()
			waitCh = make(chan time.Time)
			timerStart = false
		}
	}
}

func (l *Loader[K, V]) FetchOne(ctx context.Context, id K) (*V, error) {
	l.start.Do(l.init)
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	default:
	}

	req := loaderReq[K, V]{
		id: id,
		// We use a buffered channel so we don't have anything block if we jump out
		// of this method (e.g. for context deadline).
		ch: make(chan *entry[K, V], 1),
	}

	// Wait for context, loader shutdown, or acceptance of our request.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case l.reqCh <- req:
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}

	// Wait for context, or the loaderEntry associated with our request.
	var resp *entry[K, V]
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp = <-req.ch:
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}

	// Wait for context, or confirmation that our entry has finished loading.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-resp.done:
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	}

	return resp.data, resp.err
}
