package dataloader

import (
	"context"
	"sync"
	"time"

	"go.opencensus.io/trace"
)

type loaderConfig struct {
	FetchFunc func(context.Context, []string) ([]interface{}, error) // FetchFunc should return resources for the provided IDs (order doesn't matter).
	IDFunc    func(interface{}) string                               // Should return the unique ID for a given resource.

	Delay time.Duration // Delay before fetching pending requests.
	Max   int           // Max number of pending requests before immediate fetch (also max number of requests per-db-call).

	Name string // Name to use in traces.
}

type loaderReq struct {
	id string
	ch chan *loaderEntry
}

type loaderEntry struct {
	id   string
	done chan struct{}
	err  error
	data interface{}
}
type loader struct {
	ctx    context.Context
	cancel func()

	cfg   loaderConfig
	cache map[string]*loaderEntry

	start sync.Once

	reqCh chan loaderReq
}

func newLoader(ctx context.Context, cfg loaderConfig) *loader {
	l := &loader{cfg: cfg}
	l.ctx, l.cancel = context.WithCancel(ctx)

	return l
}

func (l *loader) Close() error {
	// ensure we don't start in the future, ensure cancel is called
	// before `start.Do` returns if it's the first call.
	l.start.Do(l.cancel)

	// always call l.cancel
	l.cancel()

	return nil
}

func (l *loader) init() {
	l.cache = make(map[string]*loaderEntry, l.cfg.Max)
	l.reqCh = make(chan loaderReq, l.cfg.Max)
	go l.loop()
}

// load will perform a batch load for a list of entries
func (l *loader) load(entries []*loaderEntry) []*loaderEntry {

	// If we need to load more than the max, call load with the max, and return the rest.
	if len(entries) > l.cfg.Max {
		l.load(entries[:l.cfg.Max])
		return entries[l.cfg.Max:]
	}

	// We need to copy the list so we don't get overwritten if other
	// batch updates are done in the background while this call is processing.
	cpy := make([]*loaderEntry, len(entries))
	copy(cpy, entries)

	go func() {
		ctx, sp := trace.StartSpan(l.ctx, "Loader.Fetch/"+l.cfg.Name)
		defer sp.End()

		// Map the entries out by ID, and collect the list of IDs
		// for the DB call.
		m := make(map[string]*loaderEntry, len(entries))
		ids := make([]string, len(entries))
		for i, e := range cpy {
			ids[i] = e.id
			m[e.id] = e
		}

		// Call fetch for everything we're loading
		res, err := l.cfg.FetchFunc(ctx, ids)
		if err != nil {
			// If the fetch failed, set all the pending entires err property to
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
			e.data = res[i]
		}

		// nil or not, all entires are now done loading, if the .data prop was not set
		// then the entry does not exist.
		for _, e := range cpy {
			close(e.done)
		}
	}()

	return entries[:0]
}

// entry will return the current entry or create a new one in the map.
// It passes the new or existing loaderEntry to the requester.
func (l *loader) entry(req loaderReq) (*loaderEntry, bool) {
	if v, ok := l.cache[req.id]; ok {
		req.ch <- v
		return v, false
	}

	e := &loaderEntry{
		id:   req.id,
		done: make(chan struct{}),
	}
	l.cache[req.id] = e
	req.ch <- e
	return e, true
}

func (l *loader) loop() {

	// timerStart tracks if the delay timer has started or not.
	var timerStart bool
	var t *time.Timer
	// waitCh by default will block indefinitely, since the timer
	// shouldn't start until the first pending request is made.
	waitCh := (<-chan time.Time)(make(chan time.Time))
	batch := make([]*loaderEntry, 0, l.cfg.Max)

	var req loaderReq
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

func (l *loader) FetchOne(ctx context.Context, id string) (interface{}, error) {
	l.start.Do(l.init)
	select {
	case <-l.ctx.Done():
		return nil, l.ctx.Err()
	default:
	}

	req := loaderReq{
		id: id,
		// We use a buffered channel so we don't have anything block if we jump out
		// of this method (e.g. for context deadline).
		ch: make(chan *loaderEntry, 1),
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
	var resp *loaderEntry
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
