package gadb

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

func NewDestV1(typeID string, args ...string) DestV1 {
	d := DestV1{
		Type: typeID,
	}

	if len(args)%2 != 0 {
		panic("args must be key-value pairs")
	}

	for i := 0; i < len(args); i += 2 {
		d.SetArg(args[i], args[i+1])
	}

	return d
}

type DestHashV1 [32]byte

func (ns DestV1) DestHash() DestHashV1 {
	data, err := json.Marshal(ns)
	if err != nil {
		panic(err)
	}

	return sha256.Sum256(data)
}

func (ns DestV1) String() string {
	return fmt.Sprintf("DestV1{Type: %s, Args: %v}", ns.Type, ns.Args)
}

type DestV1 struct {
	Args map[string]string
	Type string
}

func (ns DestV1) Equal(other DestV1) bool {
	if ns.Type != other.Type {
		return false
	}

	if len(ns.Args) != len(other.Args) {
		return false
	}

	for k, v := range ns.Args {
		if other.Args[k] != v {
			return false
		}
	}

	return true
}

func (ns DestV1) Arg(name string) string {
	if ns.Args == nil {
		return ""
	}
	return ns.Args[name]
}

func (ns *DestV1) SetArg(name, value string) {
	if ns.Args == nil {
		ns.Args = make(map[string]string)
	}
	ns.Args[name] = value
}

// Scan implements the Scanner interface.
func (ns *DestV1) Scan(value interface{}) error {
	switch v := value.(type) {
	case json.RawMessage:
		err := json.Unmarshal(v, ns)
		if err != nil {
			return err
		}
	case []byte:
		err := json.Unmarshal(v, ns)
		if err != nil {
			return err
		}
	case string:
		err := json.Unmarshal([]byte(v), ns)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported scan for DestV1 type: %T", value)
	}

	return nil
}

// Value implements the driver Valuer interface.
func (ns DestV1) Value() (interface{}, error) {
	if ns.Args == nil {
		ns.Args = map[string]string{}
	}
	data, err := json.Marshal(ns)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

type NullDestV1 struct {
	Valid  bool
	DestV1 DestV1
}

func (ns NullDestV1) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(ns.DestV1)
}

func (ns *NullDestV1) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		ns.Valid = false
		return nil
	}

	ns.Valid = true
	return json.Unmarshal(data, &ns.DestV1)
}

// Scan implements the Scanner interface.
func (ns *NullDestV1) Scan(value interface{}) error {
	if value == nil {
		ns.DestV1, ns.Valid = DestV1{}, false
		return nil
	}

	ns.Valid = true
	return ns.DestV1.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullDestV1) Value() (interface{}, error) {
	if !ns.Valid {
		return nil, nil
	}

	return ns.DestV1.Value()
}
