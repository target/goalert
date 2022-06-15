package dataloader

import (
	"context"
	"testing"
	"time"
)

func TestLoader_FetchOne(t *testing.T) {
	type example struct{ id string }
	cfg := loaderConfig[string, example]{
		Max:       10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.id },
		FetchFunc: func(context.Context, []string) ([]example, error) { return []example{{id: "foo"}}, nil },
	}
	l := newLoader(context.Background(), cfg)

	res, err := l.FetchOne(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}

	if res.id != "foo" {
		t.Errorf("got id=%s; want foo", res.id)
	}
}

func TestLoader_FetchOne_Missing(t *testing.T) {
	type example struct{ id string }
	cfg := loaderConfig[string, example]{
		Max:       10,
		Delay:     time.Millisecond,
		IDFunc:    func(v example) string { return v.id },
		FetchFunc: func(context.Context, []string) ([]example, error) { return []example{{id: "bar"}}, nil },
	}
	ctx, done := context.WithTimeout(context.Background(), time.Second)
	defer done()
	l := newLoader(ctx, cfg)

	res, err := l.FetchOne(ctx, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if res != nil {
		t.Errorf("got %T; want nil", res)
	}
}
