package mocktwilio

import (
	"net/url"
	"strings"
)

// isValidURL checks if a string is a valid http or https URL.
func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

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
