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
