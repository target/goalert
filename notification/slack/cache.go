package slack

import (
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
)

type ttlCache[K comparable, V any] struct {
	*lru.Cache
	ttl time.Duration

	// We use a mutex to protect the cache from concurrent access
	// as this is not handled by the lru package.
	//
	// See https://github.com/golang/groupcache/issues/87#issuecomment-338494548
	mx sync.Mutex

	inFlight map[K]*fillResult[V]
}

type fillResult[V any] struct {
	Value V
	Err   error
	Done  chan struct{}
}

func newTTLCache[K comparable, V any](maxEntries int, ttl time.Duration) *ttlCache[K, V] {
	return &ttlCache[K, V]{
		ttl:      ttl,
		Cache:    lru.New(maxEntries),
		inFlight: make(map[K]*fillResult[V]),
	}
}

type cacheItem[V any] struct {
	expires time.Time
	value   V
}

func (c *ttlCache[K, V]) Add(key lru.Key, value V) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c._Add(key, value)
}
func (c *ttlCache[K, V]) _Add(key lru.Key, value V) {
	c.Cache.Add(key, cacheItem[V]{
		value:   value,
		expires: time.Now().Add(c.ttl),
	})
}

func (c *ttlCache[K, V]) Get(key K) (val V, ok bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	return c._Get(key)
}
func (c *ttlCache[K, V]) _Get(key K) (val V, ok bool) {
	item, ok := c.Cache.Get(key)
	if !ok {
		return val, false
	}

	cItem := item.(cacheItem[V])
	if time.Until(cItem.expires) > 0 {
		return cItem.value, true
	}

	return val, false
}

// GetOrFill retrieves a value from the cache by key, or fills it by calling the provided function if not found or expired.
// If another goroutine is already filling the same key, it waits for that operation to complete and returns the same result.
// This prevents duplicate work and ensures only one goroutine executes the fill function for a given key at a time.
// The value is cached with the configured TTL if the fill function succeeds (returns no error).
func (c *ttlCache[K, V]) GetOrFill(key K, fn func() (V, error)) (val V, err error) {
	c.mx.Lock()
	item, ok := c._Get(key)
	if ok {
		c.mx.Unlock()
		return item, nil
	}

	res, ok := c.inFlight[key]
	if ok {
		c.mx.Unlock()
		<-res.Done
		return res.Value, res.Err
	}

	res = &fillResult[V]{
		Done: make(chan struct{}),
	}
	c.inFlight[key] = res
	c.mx.Unlock()

	val, err = fn()
	c.mx.Lock()
	delete(c.inFlight, key)
	res.Err = err
	res.Value = val
	if err == nil {
		c._Add(key, val)
	}
	close(res.Done)
	c.mx.Unlock()

	return val, err
}
