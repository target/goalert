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
