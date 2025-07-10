package dataloader

import (
	"context"
)

type batch[K comparable, V any] struct {
	ids     []K
	results []*result[K, V]
}

func (b *batch[K, V]) Add(ctx context.Context, id K) *result[K, V] {
	if b.results == nil {
		b.results = make([]*result[K, V], 0, 1)
	}
	res := &result[K, V]{id: id, done: make(chan struct{})}
	b.results = append(b.results, res)
	b.ids = append(b.ids, id)
	return res
}
