// Package dataloader provides tests for the batching and caching data loader.
package dataloader

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestFetcher_FetchOne tests basic fetching functionality with string IDs.
func TestFetcher_FetchOne(t *testing.T) {
	type example struct{ ID string }
	l := &Fetcher[string, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.ID },
		FetchFunc: func(context.Context, []string) ([]example, error) { return []example{{ID: "foo"}}, nil },
	}

	res, err := l.FetchOne(context.Background(), "foo")
	require.NoError(t, err)
	require.Equal(t, "foo", res.ID)
}

// TestFetcher_FetchOne_Int tests fetching functionality with integer IDs to ensure
// the generic type system works correctly with different comparable types.
func TestFetcher_FetchOne_Int(t *testing.T) {
	type example struct{ ID int64 }
	l := &Fetcher[int64, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) int64 { return v.ID },
		FetchFunc: func(context.Context, []int64) ([]example, error) { return []example{{ID: 42}}, nil },
	}

	res, err := l.FetchOne(context.Background(), 42)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, int64(42), res.ID)
}

// TestFetcher_FetchOne_Extra ensures that when the fetch function returns extra
// results that weren't requested, the Fetcher handles them gracefully without panicking.
func TestFetcher_FetchOne_Extra(t *testing.T) {
	type example struct{ id string }
	l := &Fetcher[string, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.id },
		FetchFunc: func(context.Context, []string) ([]example, error) { return []example{{id: "bar"}}, nil },
	}
	ctx, done := context.WithTimeout(context.Background(), time.Second)
	defer done()

	res, err := l.FetchOne(ctx, "foo")
	require.NoError(t, err)
	require.Nil(t, res)
}

// TestFetcher_MaxBatch validates that the MaxBatch setting properly limits
// the number of IDs passed to FetchFunc in a single call.
func TestFetcher_MaxBatch(t *testing.T) {
	type example struct{ ID string }

	var batchSizes []int
	l := &Fetcher[string, example]{
		MaxBatch: 3,
		Delay:    time.Millisecond,
		IDFunc:   func(v example) string { return v.ID },
		FetchFunc: func(ctx context.Context, ids []string) ([]example, error) {
			batchSizes = append(batchSizes, len(ids))
			var results []example
			for _, id := range ids {
				results = append(results, example{ID: id})
			}
			return results, nil
		},
	}

	ctx := context.Background()

	// Start 7 concurrent requests to trigger batching
	type result struct {
		res *example
		err error
	}
	results := make(chan result, 7)

	for i := 0; i < 7; i++ {
		go func(id string) {
			res, err := l.FetchOne(ctx, id)
			results <- result{res, err}
		}(string(rune('a' + i)))
	}

	// Collect all results
	for i := 0; i < 7; i++ {
		r := <-results
		require.NoError(t, r.err)
		require.NotNil(t, r.res)
	}

	// Wait a bit for batching to complete
	time.Sleep(10 * time.Millisecond)

	// Verify that no batch exceeded MaxBatch size
	for _, size := range batchSizes {
		require.LessOrEqual(t, size, 3, "Batch size should not exceed MaxBatch")
	}

	// Should have at least 3 batches (7 items with max batch size of 3)
	require.GreaterOrEqual(t, len(batchSizes), 3, "Should have created multiple batches")
}
