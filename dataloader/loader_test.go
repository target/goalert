package dataloader

import (
	"context"
	"testing"
	"time"
)

func TestLoader_FetchOne(t *testing.T) {
	type example struct{ id string }
	cfg := loaderConfig{
		Max:       10,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*example).id },
		FetchFunc: func(context.Context, []string) ([]interface{}, error) { return []interface{}{&example{id: "foo"}}, nil },
	}
	l := newLoader(context.Background(), cfg)

	res, err := l.FetchOne(context.Background(), "foo")
	if err != nil {
		t.Fatal(err)
	}

	if r, ok := res.(*example); ok {
		if r.id != "foo" {
			t.Errorf("got id=%s; want foo", r.id)
		}
	} else {
		t.Errorf("got %T; want *example", res)
	}
}

func TestLoader_FetchOne_Missing(t *testing.T) {
	type example struct{ id string }
	cfg := loaderConfig{
		Max:       10,
		Delay:     time.Millisecond,
		IDFunc:    func(v interface{}) string { return v.(*example).id },
		FetchFunc: func(context.Context, []string) ([]interface{}, error) { return []interface{}{&example{id: "bar"}}, nil },
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
