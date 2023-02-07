package slack

import (
	"time"

	"github.com/golang/groupcache/lru"
)

type ttlCache[K comparable, V any] struct {
	*lru.Cache
	ttl time.Duration
}

func newTTLCache[K comparable, V any](maxEntries int, ttl time.Duration) *ttlCache[K, V] {
	return &ttlCache[K, V]{
		ttl:   ttl,
		Cache: lru.New(maxEntries),
	}
}

type cacheItem[V any] struct {
	expires time.Time
	value   V
}

func (c *ttlCache[K, V]) Add(key lru.Key, value V) {
	c.Cache.Add(key, cacheItem[V]{
		value:   value,
		expires: time.Now().Add(c.ttl),
	})
}

func (c *ttlCache[K, V]) Get(key K) (val V, ok bool) {
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
