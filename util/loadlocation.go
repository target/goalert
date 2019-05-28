package util

import (
	"sync"
	"time"
)

var tzCache = make(map[string]*time.Location, 100)
var tzMx sync.Mutex

// LoadLocation works like time.LoadLocation but caches the result
// for the life of the process.
func LoadLocation(name string) (*time.Location, error) {
	tzMx.Lock()
	defer tzMx.Unlock()

	loc, ok := tzCache[name]
	if ok {
		return loc, nil
	}

	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil, err
	}

	tzCache[name] = loc

	return loc, nil
}
