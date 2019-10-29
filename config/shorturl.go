package config

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	uuid "github.com/satori/go.uuid"
)

// ShortURLMiddleware will issue redirects for requests to generated short URLs.
//
// Unknown/unhandled paths will be left as-is.
func ShortURLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		newPath := LongPath(req.URL.Path)
		if newPath == "" {
			next.ServeHTTP(w, req)
			return
		}

		cfg := FromContext(req.Context())
		u := *req.URL
		u.Path = newPath

		// use rawCallbackURL so we don't redirect to the same shortened URL
		http.Redirect(w, req, cfg.rawCallbackURL(u.String()).String(), http.StatusTemporaryRedirect)
	})
}

// must also be SMS-safe characters (cannot use _ for example)
const urlChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-."

var (
	alertURL         = regexp.MustCompile(`^/alerts/\d+$`)
	serviceAlertsURL = regexp.MustCompile(`^/services/[a-f0-9-]+/alerts$`)
	urlEnc           = base64.NewEncoding(urlChars).WithPadding(base64.NoPadding)
)

// ShortPath will attempt to convert a normal/long GoAlert URL into a shorter version.
//
// If unable/unknown it will return an empty string.
func ShortPath(longPath string) string {
	switch {
	case serviceAlertsURL.MatchString(longPath):
		idStr := strings.TrimPrefix(strings.TrimSuffix(longPath, "/alerts"), "/services/")
		id, err := uuid.FromString(idStr)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("/s/%s", urlEnc.EncodeToString(id.Bytes()))
	case alertURL.MatchString(longPath):
		i, err := strconv.Atoi(strings.TrimPrefix(longPath, "/alerts/"))
		if err != nil || i == 0 {
			return ""
		}
		buf := make([]byte, 8)
		n := binary.PutUvarint(buf, uint64(i))
		return fmt.Sprintf("/a/%s", urlEnc.EncodeToString(buf[:n]))
	}
	return ""
}

// LongPath will attempt to convert a shortened GoAlert URL into the original.
// If unable, it will return an empty string.
func LongPath(shortPath string) string {
	switch {
	case strings.HasPrefix(shortPath, "/a/"):
		dec, err := urlEnc.DecodeString(strings.TrimPrefix(shortPath, "/a/"))
		if err != nil {
			return ""
		}
		id, _ := binary.Uvarint(dec)
		return fmt.Sprintf("/alerts/%d", id)
	case strings.HasPrefix(shortPath, "/s/"):
		dec, err := urlEnc.DecodeString(strings.TrimPrefix(shortPath, "/s/"))
		if err != nil {
			return ""
		}
		id, err := uuid.FromBytes(dec)
		if err != nil {
			return ""
		}
		return fmt.Sprintf("/services/%s/alerts", id.String())
	}
	return ""
}
