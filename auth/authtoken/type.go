package authtoken

// Type represents the type of an authentication token.
type Type byte

// Available valid token types.
const (
	TypeUnknown Type = iota // always make the zero-value Unknown
	TypeSession
	TypeCalSub
	TypeUIK
)
