package dataloader

import (
	"context"
	"testing"
	"time"
)

func TestFetcher_FetchOne(t *testing.T) {
	type example struct{ id string }
	l := &Fetcher[string, struct{}, example]{
		MaxBatch:  10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.id },
		FetchFunc: func(context.Context, struct{}, []string) ([]example, error) { return []example{{id: "foo"}}, nil },
	}

	res, err := l.FetchOne(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}

	if res.id != "foo" {
		t.Errorf("got id=%s; want foo", res.id)
	}
}

// Ensure extra results don't cause a panic.
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

// Ensure extra results don't cause a panic.
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

// Test that FetchOneParam passes the correct parameters to the backend.
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
