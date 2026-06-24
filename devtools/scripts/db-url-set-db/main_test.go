package main

import "testing"

func TestSetDB(t *testing.T) {
	check := func(input, db, expected string) {
		t.Helper()
		actual := SetDB(input, db)
		if actual != expected {
			t.Errorf("SetDB(%q, %q) = %q; want %q", input, db, actual, expected)
		}
	}

	check("postgres://localhost:5432", "test", "postgres://localhost:5432/test")
	check("postgresql://postgres@?client_encoding=UTF8", "test", "postgresql://postgres@/test?client_encoding=UTF8")
	check("postgres://goalert@localhost:5432/goalert", "test", "postgres://goalert@localhost:5432/test")
	check("postgres://goalert@localhost:5432/goalert?sslmode=disable", "test", "postgres://goalert@localhost:5432/test?sslmode=disable")
}
