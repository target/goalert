package util

import (
	"sync"
	"time"

	"github.com/target/goalert/timezone"
	"github.com/target/goalert/validation"
)

var tzCache = make(map[string]*time.Location, 100)
var tzMx sync.Mutex

// LoadLocation works like time.LoadLocation but caches the result
// for the life of the process.
func LoadLocation(name string) (*time.Location, error) {
	tzMx.Lock()
	defer tzMx.Unlock()

	name = timezone.CanonicalZone(name)
	if name == "" {
		return nil, validation.NewFieldError("TimeZone", "unknown time zone")
	}

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
