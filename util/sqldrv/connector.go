package sqldrv

import (
	"fmt"
	"net/url"
)

// AppURL will add the application_name parameter to the provided URL.
func AppURL(urlStr, appName string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("parse db url: %w", err)
	}
	q := u.Query()
	q.Set("application_name", appName)
	q.Set("enable_seqscan", "off")
	u.RawQuery = q.Encode()
	return u.String(), nil
}
