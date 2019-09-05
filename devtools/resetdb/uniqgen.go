package main

import (
	"strings"
	"sync"
)

// uniqGen allows generating unique string values.
type uniqGen struct {
	m  map[strScope]int
	mx sync.Mutex
}

type strScope struct {
	scope string
	value string
}

func newGen() *uniqGen {
	return &uniqGen{
		m: make(map[strScope]int),
	}
}

// PickOne will return a random item from a slice, and will not return the
// same value twice.
func (g *uniqGen) PickOne(s []string) string {
	return g.Gen(func() string { return sample(s) })
}

// Gen will call `fn` until a new value is returned.
func (g *uniqGen) Gen(fn func() string, scope ...string) string { return g.GenN(1, fn, scope...) }

// Gen will call `fn` until a value that has been used less than N is returned.
func (g *uniqGen) GenN(n int, fn func() string, scope ...string) string {
	g.mx.Lock()
	defer g.mx.Unlock()
	scopeVal := strings.Join(scope, "|")
	for {
		scope := strScope{
			scope: scopeVal,
			value: fn(),
		}
		if g.m[scope] >= n {
			continue
		}
		g.m[scope]++
		return scope.value
	}
}
