package notification

import (
	"crypto/sha256"
	"sort"
)

type DestV2 struct {
	Type DestTypeV2
	Args DestArgs
}

type (
	DestTypeV2 string
	DestArgs   map[string]string
)

// DestHash is a comparable-type for distinguishing unique destination values.
type DestHash [32]byte

// DestType returns the type of the destination.
func (d DestV2) DestType() DestTypeV2 { return d.Type }

// DestArg will return the value of the named argument for the destination.
func (d DestV2) DestArg(name string) string {
	if d.Args == nil {
		return ""
	}

	return d.Args[name]
}

// DestHash will return a unique & stable hash for the destination.
func (d DestV2) DestHash() DestHash {
	keys := make([]string, 0, len(d.Args))
	for k := range d.Args {
		keys = append(keys, k)
	}

	// sort keys
	sort.Strings(keys)

	h := sha256.New()
	h.Write([]byte(d.Type + "\n"))
	for _, k := range keys {
		h.Write([]byte(k + "\n"))
		h.Write([]byte(d.Args[k] + "\n"))
	}

	return DestHash(h.Sum(nil))
}
