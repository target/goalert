package util

import (
	"net/url"
	"strings"
)

// JoinURL will join a base URL and suffix, taking care to preserve and merge query parameters.
func JoinURL(base, suffix string) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	su, err := url.Parse(suffix)
	if err != nil {
		return "", err
	}

	u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(su.Path, "/")

	v := u.Query()
	for name := range su.Query() {
		v.Set(name, su.Query().Get(name))
	}
	u.RawQuery = v.Encode()

	return u.String(), nil
}
