package util

import (
	"context"
	"errors"
	"testing"
)

func TestCache(t *testing.T) {
	c := NewContextCache()

	c1, can1 := context.WithCancel(cacheableContext(context.Background()))
	c2, can2 := context.WithCancel(cacheableContext(context.Background()))
	c3 := context.Background()
	defer can1()
	defer can2()

	testLoad := func(ctx context.Context, key, val string) {
		t.Run("Load", func(t *testing.T) {
			v := c.Load(ctx, key)
			if val == "" && v != nil {
				t.Fatalf("load unknown key: got %+v; want nil", v)
			}
			if val == "" {
				return
			}

			if v.(string) != val {
				t.Errorf("load stored string: got '%s'; want '%s'", v.(string), val)
			}
		})
	}

	// non-wrapped contexts don't cache anything
	testLoad(c3, "foo", "")
	c.Store(c3, "foo", "bar")
	testLoad(c3, "foo", "")

	testLoad(c1, "foo", "")
	c.Store(c1, "foo", "bar")
	testLoad(c1, "foo", "bar")
	c.Store(c1, "foo", "baz")
	testLoad(c1, "foo", "baz")
	testLoad(c2, "foo", "") // cache should be per-cacheableContext

	type k string

	// cache should follow all child contexts
	c1Sub := context.WithValue(c1, k("test"), nil)
	testLoad(c1Sub, "foo", "baz")
	c.Store(c1Sub, "foo", "blah")
	testLoad(c1Sub, "foo", "blah")
	testLoad(c1, "foo", "blah")

	testLS := func(key, retrVal string, retrErr error, expVal string, expErr error) {
		t.Run("LoadOrStore", func(t *testing.T) {
			v, err := c.LoadOrStore(c1, key, func() (interface{}, error) {
				return retrVal, retrErr
			})
			if err != expErr {
				t.Errorf("err was %+v; want %+v", err, expErr)
			}
			m, _ := v.(string)
			if m != expVal {
				t.Errorf("val was '%s'; want '%s'", m, expVal)
			}
		})
	}
	e := errors.New("broke")
	data := []struct {
		key  string
		retV string
		retE error
		expV string
		expE error
	}{
		{"foo", "a", nil, "bar", nil},
		{"foo", "a", e, "bar", nil},
		{"bin", "baz", e, "baz", e},
		{"bin", "foo", e, "foo", e}, //don't store on err
		{"bin", "foo", nil, "foo", nil},
		{"bin", "fwah", e, "foo", nil},
	}
	c.Store(c1, "foo", "bar")
	for _, d := range data {
		testLS(d.key, d.retV, d.retE, d.expV, d.expE)
	}

}
