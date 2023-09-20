package apikey

import (
	"context"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
)

type lastUsedCache struct {
	lru *lru.Cache

	mx         sync.Mutex
	updateFunc func(ctx context.Context, id uuid.UUID, ua, ip string) error
}

func newLastUsedCache(max int, updateFunc func(ctx context.Context, id uuid.UUID, ua, ip string) error) *lastUsedCache {
	return &lastUsedCache{
		lru:        lru.New(max),
		updateFunc: updateFunc,
	}
}
func (c *lastUsedCache) RecordUsage(ctx context.Context, id uuid.UUID, ua, ip string) error {
	c.mx.Lock()
	defer c.mx.Unlock()
	if t, ok := c.lru.Get(id); ok && time.Since(t.(time.Time)) < time.Minute {
		return nil
	}

	c.lru.Add(id, time.Now())
	return c.updateFunc(ctx, id, ua, ip)
}
