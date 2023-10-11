package main

import (
	"strings"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
)

// uniqGen allows generating unique string values.
type uniqGen struct {
	f  *gofakeit.Faker
	m  map[strScope]int
	mx sync.Mutex
}

type strScope struct {
	scope string
	value string
}

func newGen(f *gofakeit.Faker) *uniqGen {
	return &uniqGen{
		f: f,
		m: make(map[strScope]int, 100000),
	}
}

// PickOne will return a random item from a slice, and will not return the
// same value twice.
func (g *uniqGen) PickOne(s []string) string {
	return g.Gen(func() string { return g.f.RandomString(s) })
}

// Gen will call `fn` until a unique value is returned.
func (g *uniqGen) Gen(fn func() string, scope ...string) string { return g.GenN(1, fn, scope...) }

// Gen will call `fn` until a value that has been returned less than N is provided.
func (g *uniqGen) GenN(n int, fn func() string, scope ...string) string {
	g.mx.Lock()
	defer g.mx.Unlock()
	scopeVal := strings.Join(scope, "|")
	var i int
	for {
		if i > 5 {
			panic("aborted after 5 tries")
		}
		scope := strScope{
			scope: scopeVal,
			value: fn(),
		}
		if g.m[scope] >= n {
			i++
			continue
		}
		g.m[scope]++
		return scope.value
	}
}
