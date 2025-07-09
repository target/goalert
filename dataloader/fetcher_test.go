// Package dataloader provides tests for the batching and caching data loader.
package dataloader

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// TestFetcher_FetchOne tests basic fetching functionality with string IDs.
func TestFetcher_FetchOne(t *testing.T) {
	type example struct{ ID string }
	l := &Fetcher[string, struct{}, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.ID },
		FetchFunc: func(context.Context, struct{}, []string) ([]example, error) { return []example{{ID: "foo"}}, nil },
	}

	res, err := l.FetchOne(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}

	if res.ID != "foo" {
		t.Errorf("got id=%s; want foo", res.ID)
	}
}

// TestFetcher_FetchOne_UUID_String tests that UUIDs and strings can be used together. This is important because as we transition from strings to UUIDs, it's commong for code to use a UUID while FetchOne is given a string, and vice versa.
func TestFetcher_FetchOne_UUID_String(t *testing.T) {
	type exampleUUID struct{ ID uuid.UUID }
	type exampleString struct{ ID string }
	testID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	unknownID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	t.Run("FetchOne with UUID arg with string struct ID", func(t *testing.T) {
		l := &Fetcher[uuid.UUID, struct{}, exampleString]{
			MaxBatch: 10,
			Delay:    time.Millisecond,
			FetchFunc: func(context.Context, struct{}, []uuid.UUID) ([]exampleString, error) {
				return []exampleString{{ID: testID.String()}}, nil
			},
		}

		res, err := l.FetchOne(context.Background(), testID)
		require.NoError(t, err)
		require.Equal(t, testID.String(), res.ID, "expected ID to match testID string")

		res, err = l.FetchOne(context.Background(), unknownID)
		require.NoError(t, err)
		require.Nil(t, res, "expected nil result for unknown ID")
	})

	t.Run("FetchOne with string arg with UUID struct ID", func(t *testing.T) {
		l := &Fetcher[string, struct{}, exampleUUID]{
			MaxBatch: 10,
			Delay:    time.Millisecond,
			IDFunc:   func(v exampleUUID) string { return v.ID.String() },
			FetchFunc: func(context.Context, struct{}, []string) ([]exampleUUID, error) {
				return []exampleUUID{{ID: testID}}, nil
			},
		}

		res, err := l.FetchOne(context.Background(), testID.String())
		require.NoError(t, err)
		require.Equal(t, testID, res.ID, "expected ID to match testID")

		res, err = l.FetchOne(context.Background(), unknownID.String())
		require.NoError(t, err)
		require.Nil(t, res, "expected nil result for unknown ID")
	})
}

// TestFetcher_FetchOne_Int tests fetching functionality with integer IDs to ensure
// the generic type system works correctly with different comparable types.
func TestFetcher_FetchOne_Int(t *testing.T) {
	type example struct{ ID int }
	l := &Fetcher[int, struct{}, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		FetchFunc: func(context.Context, struct{}, []int) ([]example, error) { return []example{{ID: 2}}, nil },
	}

	res, err := l.FetchOne(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	require.NotNil(t, res)

	if res.ID != 2 {
		t.Errorf("got id=%d; want 2", res.ID)
	}
}

// TestFetcher_FetchOne_Extra ensures that when the fetch function returns extra
// results that weren't requested, the Fetcher handles them gracefully without panicking.
func TestFetcher_FetchOne_Extra(t *testing.T) {
	type example struct{ id string }
	l := &Fetcher[string, struct{}, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.id },
		FetchFunc: func(context.Context, struct{}, []string) ([]example, error) { return []example{{id: "bar"}}, nil },
	}
	ctx, done := context.WithTimeout(context.Background(), time.Second)
	defer done()

	res, err := l.FetchOne(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if res != nil {
		t.Errorf("got %T; want nil", res)
	}
}

// TestFetcher_FetchOne_Missing tests that when the fetch function doesn't return
// a requested ID, the Fetcher correctly returns nil without error.
func TestFetcher_FetchOne_Missing(t *testing.T) {
	type example struct{ id string }
	l := &Fetcher[string, struct{}, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.id },
		FetchFunc: func(context.Context, struct{}, []string) ([]example, error) { return nil, nil },
	}
	ctx, done := context.WithTimeout(context.Background(), time.Second)
	defer done()

	res, err := l.FetchOne(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if res != nil {
		t.Errorf("got %T; want nil", res)
	}
}

// TestFetcher_FetchOneParam_ValidatesParams tests that FetchOneParam correctly
// passes the specified parameters to the fetch function, enabling parameter-specific
// batching and filtering.
func TestFetcher_FetchOneParam_ValidatesParams(t *testing.T) {
	type example struct{ id string }
	type params struct{ filter string }

	var capturedParam params
	var capturedIDs []string

	l := &Fetcher[string, params, example]{
		MaxBatch: 10,
		Delay:    time.Millisecond,
		IDFunc:   func(v example) string { return v.id },
		FetchFunc: func(ctx context.Context, param params, ids []string) ([]example, error) {
			capturedParam = param
			capturedIDs = ids
			return []example{{id: "foo"}}, nil
		},
	}

	ctx, done := context.WithTimeout(context.Background(), time.Second)
	defer done()

	expectedParam := params{filter: "active"}
	res, err := l.FetchOneParam(ctx, "foo", expectedParam)
	if err != nil {
		t.Fatal(err)
	}

	if res == nil || res.id != "foo" {
		t.Errorf("got %v; want example with id 'foo'", res)
	}

	if capturedParam != expectedParam {
		t.Errorf("got param %+v; want %+v", capturedParam, expectedParam)
	}

	if len(capturedIDs) != 1 || capturedIDs[0] != "foo" {
		t.Errorf("got IDs %v; want ['foo']", capturedIDs)
	}
}
