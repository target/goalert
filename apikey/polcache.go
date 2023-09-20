package apikey

import (
	"context"
	"database/sql"
	"errors"
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/google/uuid"
	"github.com/target/goalert/gadb"
)

// polCache handles caching of policyInfo objects, as well as negative caching
// of invalid keys.
type polCache struct {
	lru *lru.Cache
	neg *lru.Cache
	mx  sync.Mutex

	cfg polCacheConfig
}

type polCacheConfig struct {
	FillFunc func(context.Context, uuid.UUID) (*policyInfo, bool, error)
	Verify   func(context.Context, uuid.UUID) (bool, error)
	MaxSize  int
}

func newPolCache(cfg polCacheConfig) *polCache {
	return &polCache{
		lru: lru.New(cfg.MaxSize),
		neg: lru.New(cfg.MaxSize),
		cfg: cfg,
	}
}

// Revoke will add the key to the negative cache.
func (c *polCache) Revoke(ctx context.Context, key uuid.UUID) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.neg.Add(key, nil)
	c.lru.Remove(key)

	return nil
}

// Get will return the policyInfo for the given key.
//
// If the key is in the cache, it will be verified before returning.
//
// If it is not in the cache, it will be fetched and added to the cache.
//
// If either the key is invalid or the policy is invalid, the key will be
// added to the negative cache.
func (c *polCache) Get(ctx context.Context, key uuid.UUID) (value *policyInfo, ok bool, err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if _, ok := c.neg.Get(key); ok {
		return value, false, nil
	}

	if v, ok := c.lru.Get(key); ok {
		// Check if the key is still valid before returning it,
		// if it is not valid, we can remove it from the cache.
		isValid, err := c.cfg.Verify(ctx, key)
		if err != nil {
			return value, false, err
		}

		// Since each key has a unique ID and is signed, we can
		// safely assume that an invalid key will always be invalid
		// and can be negatively cached.
		if !isValid {
			c.neg.Add(key, nil)
			c.lru.Remove(key)
			return value, false, nil
		}

		return v.(*policyInfo), true, nil
	}

	// If the key is not in the cache, we need to fetch it,
	// and add it to the cache. We can safely assume that
	// the key is valid when returned from the FillFunc.
	value, isValid, err := c.cfg.FillFunc(ctx, key)
	if err != nil {
		return value, false, err
	}
	if !isValid {
		c.neg.Add(key, nil)
		return value, false, nil
	}

	c.lru.Add(key, value)
	return value, true, nil
}

func (s *Store) _verifyPolicyID(ctx context.Context, id uuid.UUID) (bool, error) {
	valid, err := gadb.New(s.db).APIKeyAuthCheck(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return valid, nil
}
