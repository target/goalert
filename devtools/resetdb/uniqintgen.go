package main

import (
	"math/rand"
	"strings"
	"sync"
)

// uniqIntGen works like uniqGen but returns integers.
type uniqIntGen struct {
	m  map[intScope]bool
	mx sync.Mutex
}

type intScope struct {
	scope string
	value int
}

func newUniqIntGen() *uniqIntGen {
	return &uniqIntGen{
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
			value: rand.Intn(n),
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
