package validate

import (
	"github.com/target/goalert/validation"
	"net/url"
)

// URL will validate a URL, returning a FieldError
// if invalid.
func URL(fname, urlStr string) error {
	if _, err := url.Parse(urlStr); err != nil {
		return validation.NewFieldError(fname, "must be a valid URL: "+err.Error())
	}
	return nil
}

// AbsoluteURL will validate that a URL is valid and contains
// a scheme and host.
func AbsoluteURL(fname, urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return validation.NewFieldError(fname, "must be a valid URL: "+err.Error())
	}
	if u.Scheme == "" {
		return validation.NewFieldError(fname, "scheme is required for URL")
	}
	if u.Host == "" {
		return validation.NewFieldError(fname, "host is required for URL")
	}
	return nil
}
