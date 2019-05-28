package util

import (
	"context"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

type cacheableKey string

const cacheableKeyID = cacheableKey("cache-id")

func cacheableContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, cacheableKeyID, uuid.NewV4().String())
}

// WrapCacheableContext will make all request contexts cacheable, to be used with
// a ContextCache.
func WrapCacheableContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		next.ServeHTTP(w, req.WithContext(cacheableContext(req.Context())))
	})
}

// A ContextCache is used to cache and load values on a per-context basis.
// No values will be stored unless the context.Context has passed through
// WrapCacheableContext.
type ContextCache interface {
	Load(ctx context.Context, id string) interface{}
	Store(ctx context.Context, id string, value interface{})
	LoadOrStore(context.Context, string, func() (interface{}, error)) (interface{}, error)
}

type cacheRegister struct {
	cID string

	id    string
	value interface{}

	done <-chan struct{}
}
type cacheRequest struct {
	cID string
	id  string
	ch  chan interface{}
}

type chanCache struct {
	setCh   chan *cacheRegister
	getCh   chan *cacheRequest
	cleanCh chan string

	cache map[string]map[string]interface{}
}

// NewContextCache creates a new ContextCache
func NewContextCache() ContextCache {
	return newChanCache()
}

func newChanCache() *chanCache {
	c := &chanCache{
		setCh:   make(chan *cacheRegister),
		getCh:   make(chan *cacheRequest),
		cleanCh: make(chan string),

		cache: make(map[string]map[string]interface{}, 4000),
	}
	go c.loop()
	return c
}
func (c *chanCache) cleanup(cid string, ch <-chan struct{}) {
	<-ch
	c.cleanCh <- cid
}
func (c *chanCache) loop() {
	for {
		select {
		case cid := <-c.cleanCh:
			delete(c.cache, cid)
		case reg := <-c.setCh:
			m := c.cache[reg.cID]
			if m == nil {
				m = make(map[string]interface{})
				c.cache[reg.cID] = m
				go c.cleanup(reg.cID, reg.done)
			}
			m[reg.id] = reg.value
		case req := <-c.getCh:
			m := c.cache[req.cID]
			if m == nil {
				req.ch <- nil
			}
			req.ch <- m[req.id]
		}
	}
}

func (c *chanCache) Load(ctx context.Context, id string) interface{} {
	cID, ok := ctx.Value(cacheableKeyID).(string)
	if !ok {
		return nil
	}

	ch := make(chan interface{}, 1)
	c.getCh <- &cacheRequest{cID: cID, ch: ch, id: id}

	return <-ch
}
func (c *chanCache) LoadOrStore(ctx context.Context, id string, fn func() (interface{}, error)) (interface{}, error) {
	v := c.Load(ctx, id)
	if v != nil {
		return v, nil
	}

	v, err := fn()
	if err == nil && v != nil {
		c.Store(ctx, id, v)
	}
	return v, err
}
func (c *chanCache) Store(ctx context.Context, id string, val interface{}) {
	cID, ok := ctx.Value(cacheableKeyID).(string)
	if !ok {
		return
	}

	c.setCh <- &cacheRegister{cID: cID, id: id, value: val, done: ctx.Done()}
}
