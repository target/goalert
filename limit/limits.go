package limit

// Limits contains the current value of all configurable limits.
type Limits map[ID]int

// Max returns the current max value of the limit with the given ID.
func (l Limits) Max(id ID) int {
	v, ok := l[id]
	if !ok {
		return -1
	}
	return v
}
