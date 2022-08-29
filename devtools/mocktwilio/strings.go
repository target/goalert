package mocktwilio

import (
	"net/url"
	"path"
	"strings"
)

func relURL(oldURL, newURL string) string {
	if newURL == "" {
		return oldURL
	}

	uNew, err := url.Parse(newURL)
	if err != nil {
		return ""
	}
	if uNew.Scheme != "" {
		return newURL
	}
	uOld, err := url.Parse(oldURL)
	if err != nil {
		return ""
	}
	uOld.RawQuery = uNew.RawQuery
	// use `/base` as a temporary prefix to validate the new path doesn't back up past the root (Twilio considers this an error)
	if strings.HasPrefix(uNew.Path, "/") {
		uOld.Path = path.Join("/base", uNew.Path)
	} else {
		uOld.Path = path.Join("/base", path.Dir(uOld.Path), uNew.Path)
	}
	if !strings.HasPrefix(uOld.Path, "/base") {
		return ""
	}
	uOld.Path = strings.TrimPrefix(uOld.Path, "/base")

	return uOld.String()
}

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
