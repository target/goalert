package log

import (
	"context"
	"testing"
)

func TestWithField(t *testing.T) {
	ctx := context.Background()

	foo := WithField(ctx, "foo", "bar")
	bin := WithField(ctx, "bin", "baz")

	m := ContextFields(ctx)
	if len(m) != 0 {
		t.Errorf("no. fields for background ctx is %d; want 0", len(m))
	}

	m = ContextFields(foo)
	if len(m) != 1 {
		t.Errorf("no. fields for background ctx is %d; want 1", len(m))
	}
	val, _ := m["foo"].(string)
	if val != "bar" {
		t.Errorf("foo = %s; want bar", val)
	}

	m = ContextFields(bin)
	if len(m) != 1 {
		t.Errorf("no. fields for background ctx is %d; want 1", len(m))
	}
	val, _ = m["bin"].(string)
	if val != "baz" {
		t.Errorf("bin = %s; want baz", val)
	}

	foo2 := WithField(foo, "foo2", "bar2")

	m = ContextFields(foo)
	if len(m) != 1 {
		t.Errorf("no. fields for background ctx is %d; want 1", len(m))
	}
	m = ContextFields(foo2)
	if len(m) != 2 {
		t.Errorf("no. fields for background ctx is %d; want 2", len(m))
	}
	val, _ = m["foo2"].(string)
	if val != "bar2" {
		t.Errorf("foo2 = %s; want bar2", val)
	}

	foo3 := WithField(foo, "foo", "blah")
	m = ContextFields(foo3)
	if len(m) != 1 {
		t.Errorf("no. fields for background ctx is %d; want 1", len(m))
	}
	val, _ = m["foo"].(string)
	if val != "blah" {
		t.Errorf("foo = %s; want blah", val)
	}

}
