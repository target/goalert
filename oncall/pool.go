package oncall

import (
	"sync"
)

var (
	activeMapPool = &sync.Pool{
		New: func() interface{} { return make(map[string]struct{}, 20) },
	}
	overrideMapPool = &sync.Pool{
		New: func() interface{} { return make(map[string]string, 20) },
	}
	shiftMapPool = &sync.Pool{
		New: func() interface{} { return make(map[string]*Shift, 20) },
	}
)

func getShiftMap() map[string]*Shift { return shiftMapPool.Get().(map[string]*Shift) }
func putShiftMap(m map[string]*Shift) {
	for k := range m {
		delete(m, k)
	}
	shiftMapPool.Put(m)
}

func getActiveMap() map[string]struct{} { return activeMapPool.Get().(map[string]struct{}) }
func putActiveMap(m map[string]struct{}) {
	for k := range m {
		delete(m, k)
	}
	activeMapPool.Put(m)
}

func getOverrideMap() map[string]string { return overrideMapPool.Get().(map[string]string) }
func putOverrideMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
	overrideMapPool.Put(m)
}
