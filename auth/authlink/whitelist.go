package authlink

import (
	"context"
	"errors"
	"sync"
)

type Whitelist struct {
	mx sync.Mutex
	m  map[string]chan bool
}

func NewWhitelist() *Whitelist {
	return &Whitelist{m: make(map[string]chan bool)}
}

func (w *Whitelist) Set(ids []string) {
	w.mx.Lock()
	defer w.mx.Unlock()
	oldMap := w.m
	w.m = make(map[string]chan bool)
	for _, id := range ids {
		ch := oldMap[id]
		if ch == nil {
			ch = make(chan bool, 1)
			ch <- true
		} else {
			delete(oldMap, id)
		}
		w.m[id] = ch
	}

	for _, ch := range oldMap {
		close(ch)
	}
}

var (
	ErrBadID = errors.New("no value with that ID")
)

func (w *Whitelist) LockRemove(ctx context.Context, id string, fn func(context.Context) error) error {
	w.mx.Lock()
	ch := w.m[id]
	w.mx.Unlock()
	if ch == nil {
		return ErrBadID
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case gotLock := <-ch:
		if !gotLock {
			return ErrBadID
		}
	}

	err := fn(ctx)
	if err != nil {
		w.mx.Lock()
		select {
		case <-ch:
			// closed
		default:
			ch <- true
		}
		w.mx.Unlock()
		return err
	}

	w.mx.Lock()
	select {
	case <-ch:
		// closed
	default:
		close(ch)
	}
	w.mx.Unlock()

	return nil
}
