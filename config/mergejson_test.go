package config

import "testing"

func TestMergeJSON(t *testing.T) {
	const orig = `{"foo": "bar", "bin": "baz", "d": {"e": "f"}}`
	const add = `{"bin":"", "ok": "then", "a":{"b":"c"}, "d":{"e":"g"}}`
	const exp = `{"a":{"b":"c"},"bin":"","d":{"e":"g"},"foo":"bar","ok":"then"}`

	data, err := mergeJSON([]byte(orig), []byte(add))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != exp {
		t.Fatalf("got '%s'; want '%s'", string(data), exp)
	}
}
