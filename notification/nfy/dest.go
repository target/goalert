package nfy

import (
	"crypto/sha256"
	"sort"

	"github.com/target/goalert/gadb"
)

type Dest struct {
	Type DestType
	Args DestArgs
}

func FromSQL(d gadb.DestV1) Dest {
	return Dest{
		Type: DestType(d.Type),
		Args: DestArgs(d.Args),
	}
}

// NewDest creates a new destination with the given type and arguments.
func NewDest(t DestType, args ...string) Dest {
	if len(args)%2 != 0 {
		panic("args must be key-value pairs")
	}

	dest := Dest{Type: t, Args: make(DestArgs)}
	for i := 0; i < len(args); i += 2 {
		dest.Args[args[i]] = args[i+1]
	}

	return dest
}

type (
	DestType string
	DestArgs map[string]string
)

// String will return the string representation of the destination type.
func (t DestType) String() string { return string(t) }

// DestHash is a comparable-type for distinguishing unique destination values.
type DestHash [32]byte

func (d Dest) DestType() DestType {
	return d.Type
}

// Arg will return the value of the named argument for the destination.
func (d Dest) DestArg(name string) string {
	if d.Args == nil {
		return ""
	}

	return d.Args[name]
}

// Hash will return a unique & stable hash for the destination.
func (d Dest) DestHash() DestHash {
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
