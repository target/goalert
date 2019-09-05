package main

import "math/rand"

// sample will return a single value at random from a slice.
func sample(s []string) string {
	return s[rand.Intn(len(s))]
}
