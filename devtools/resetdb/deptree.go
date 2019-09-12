package main

import "sync"

// depTree manages tasks with dependencies, allowing them to execute in parallel
// as quickly as possible.
type depTree struct {
	done map[string]bool
	mx   sync.Mutex
	ch   chan struct{}
}

// newDepTree creates a new depTree that can be used to managed parallel tasks.
func newDepTree() *depTree {
	return &depTree{
		done: make(map[string]bool),
		ch:   make(chan struct{}),
	}
}

// Start registers a new task ID and switches it to a "busy" state.
func (dt *depTree) Start(id string) {
	dt.mx.Lock()
	defer dt.mx.Unlock()
	dt.done[id] = false
}

// Done switches the provided task ID to a "done" state. It is registered if not already.
func (dt *depTree) Done(id string) {
	dt.mx.Lock()
	defer dt.mx.Unlock()
	dt.done[id] = true
	close(dt.ch)
	dt.ch = make(chan struct{})
}

// WaitFor will block until all provided ids are registered, and reported as "done".
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

// Wait will block until all registered IDs are reported as "done".
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
