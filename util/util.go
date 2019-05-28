package util

import (
	"net/url"
	"strings"
)

// JoinURL will join a base URL and suffix, taking care to preserve and merge query parameters.
func JoinURL(base, suffix string) (string, error) {
	bu, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	su, err := url.Parse(suffix)
	if err != nil {
		return "", err
	}

	bu.Path = strings.TrimSuffix(bu.Path, "/") + "/" + strings.TrimPrefix(su.Path, "/")

	v := bu.Query()
	for name := range su.Query() {
		v.Set(name, su.Query().Get(name))
	}
	bu.RawQuery = v.Encode()

	return bu.String(), nil
}
