package expflag

import "sort"

type Flag string

const (
	Example   Flag = "example"
	DestTypes Flag = "dest-types"
	UnivKeys  Flag = "univ-keys"
)

var desc = map[Flag]string{
	Example:   "An example experimental flag to demonstrate usage.",
	DestTypes: "Generic destination type API.",
	UnivKeys:  "Universal integration key support.",
}

// AllFlags returns a slice of all experimental flags sorted by name.
func AllFlags() []Flag {
	var result []Flag
	for k := range desc {
		result = append(result, k)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

// Description returns the description of the given flag.
func Description(f Flag) string {
	return desc[f]
}
