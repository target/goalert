package mocktwilio

import "strings"

func toLowerSlice(s []string) []string {
	for i, a := range s {
		s[i] = strings.ToLower(a)
	}
	return s
}

func containsAll(body string, vals []string) bool {
	body = strings.ToLower(body)
	for _, a := range toLowerSlice(vals) {
		if !strings.Contains(body, a) {
			return false
		}
	}

	return true
}
