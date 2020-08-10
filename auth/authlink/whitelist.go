package authlink

import (
	"context"
	"errors"
	"sync"
)

type Whitelist struct {
	mx sync.Mutex
	m  map[string]*wlEntry
}

func NewWhitelist() *Whitelist {
	return &Whitelist{m: make(map[string]*wlEntry)}
}

type wlEntry struct {
	wait    chan bool
	invalid bool
}

func (w *Whitelist) Set(ids []string) {
	w.mx.Lock()
	defer w.mx.Unlock()
	oldMap := w.m
	w.m = make(map[string]*wlEntry)
	for _, id := range ids {
		e := oldMap[id]
		if e == nil {
			e = &wlEntry{wait: make(chan bool, 1)}
			e.wait <- true
		} else {
			delete(oldMap, id)
		}
		w.m[id] = e
	}

	for _, e := range oldMap {
		if e.invalid {
			continue
		}
		e.invalid = true
		close(e.wait)
	}
}

var (
	ErrBadID = errors.New("no value with that ID")
)

func (w *Whitelist) LockRemove(ctx context.Context, id string, fn func(context.Context) error) error {
	w.mx.Lock()
	e := w.m[id]
	valid := e != nil && !e.invalid
	w.mx.Unlock()
	if !valid {
		return ErrBadID
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case gotLock := <-e.wait:
		if !gotLock {
			return ErrBadID
		}
	}

	err := fn(ctx)
	if err != nil {
		w.mx.Lock()
		defer w.mx.Unlock()

		if e.invalid {
			return err
		}
		if err == ErrBadID {
			e.invalid = true
			close(e.wait)
			return err
		}

		// other error, allow retry
		e.wait <- true

		return err
	}

	w.mx.Lock()
	defer w.mx.Unlock()
	if e.invalid {
		return nil
	}

	e.invalid = true
	close(e.wait)

	return nil
}
