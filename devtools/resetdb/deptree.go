package main

import "sync"

type depTree struct {
	done map[string]bool
	mx   sync.Mutex
	ch   chan struct{}
}

func newDepTree() *depTree {
	return &depTree{
		done: make(map[string]bool),
		ch:   make(chan struct{}),
	}
}

func (dt *depTree) Start(id string) {
	dt.mx.Lock()
	defer dt.mx.Unlock()
	dt.done[id] = false
}
func (dt *depTree) Done(id string) {
	dt.mx.Lock()
	defer dt.mx.Unlock()
	dt.done[id] = true
	close(dt.ch)
	dt.ch = make(chan struct{})
}
func (dt *depTree) WaitFor(ids ...string) {
	for {
		isDone := true
		dt.mx.Lock()
		for _, id := range ids {
			if !dt.done[id] {
				isDone = false
				break
			}
		}
		ch := dt.ch
		dt.mx.Unlock()
		if isDone {
			return
		}
		<-ch
	}
}
func (dt *depTree) Wait() {
	for {
		isDone := true
		dt.mx.Lock()
		for _, val := range dt.done {
			if !val {
				isDone = false
				break
			}
		}
		ch := dt.ch
		dt.mx.Unlock()
		if isDone {
			return
		}
		<-ch
	}
}
