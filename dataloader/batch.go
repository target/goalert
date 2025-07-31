package dataloader

import (
	"context"
)

// batch represents a collection of requests that will be executed together.
// It maintains both the IDs to fetch and the result channels to notify when complete.
type batch[K comparable, V any] struct {
	ids     []K              // IDs to fetch in this batch
	results []*result[K, V]  // Result channels for each request
}

// Add appends a new request to the batch and returns a result channel.
// The caller can wait on the result channel to receive the fetched value.
func (b *batch[K, V]) Add(ctx context.Context, id K) *result[K, V] {
	if b.results == nil {
		b.results = make([]*result[K, V], 0, 1)
	}
	res := &result[K, V]{id: id, done: make(chan struct{})}
	b.results = append(b.results, res)
	b.ids = append(b.ids, id)
	return res
}
