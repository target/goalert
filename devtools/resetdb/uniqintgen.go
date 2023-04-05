package main

import (
	"strings"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
)

// uniqIntGen works like uniqGen but returns integers.
type uniqIntGen struct {
	f  *gofakeit.Faker
	m  map[intScope]bool
	mx sync.Mutex
}

type intScope struct {
	scope string
	value int
}

func newUniqIntGen(f *gofakeit.Faker) *uniqIntGen {
	return &uniqIntGen{
		f: f,
		m: make(map[intScope]bool),
	}
}

// Gen will return a random value from 0 to n (non-inclusive).
//
// It will always return a unique value.
func (g *uniqIntGen) Gen(n int, scope ...string) int {
	g.mx.Lock()
	defer g.mx.Unlock()
	scopeVal := strings.Join(scope, "|")
	var i int
	for {
		if i > 5 {
			panic("aborted after 5 tries")
		}
		scope := intScope{
			value: g.f.IntRange(0, n-1),
			scope: scopeVal,
		}
		if g.m[scope] {
			i++
			continue
		}
		g.m[scope] = true
		return scope.value
	}
}
