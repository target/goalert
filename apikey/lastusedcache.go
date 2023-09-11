package apikey

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
	"github.com/target/goalert/validation/validate"
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

func (s *Store) _updateLastUsed(ctx context.Context, id uuid.UUID, ua, ip string) error {
	ua = validate.SanitizeText(ua, 1024)
	ip, _, _ = net.SplitHostPort(ip)
	ip = validate.SanitizeText(ip, 255)
	params := gadb.APIKeyRecordUsageParams{
		KeyID:     id,
		UserAgent: ua,
	}
	params.IpAddress.IPNet.IP = net.ParseIP(ip)
	params.IpAddress.IPNet.Mask = net.CIDRMask(32, 32)
	if params.IpAddress.IPNet.IP != nil {
		params.IpAddress.Valid = true
	}
	return gadb.New(s.db).APIKeyRecordUsage(ctx, params)
}
