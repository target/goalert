package slack

import (
	"time"

	"github.com/golang/groupcache/lru"
)

type ttlCache struct {
	*lru.Cache
	ttl time.Duration
}

func newTTLCache(maxEntries int, ttl time.Duration) *ttlCache {
	return &ttlCache{
		ttl:   ttl,
		Cache: lru.New(maxEntries),
	}
}

type cacheItem struct {
	expires time.Time
	value   interface{}
}

func (c *ttlCache) Add(key lru.Key, value interface{}) {
	c.Cache.Add(key, cacheItem{
		value:   value,
		expires: time.Now().Add(c.ttl),
	})
}

func (c *ttlCache) Get(key lru.Key) (interface{}, bool) {
	item, ok := c.Cache.Get(key)
	if !ok {
		return nil, false
	}
	cItem := item.(cacheItem)
	if time.Until(cItem.expires) > 0 {
		return cItem.value, true
	}
	return nil, false
}
