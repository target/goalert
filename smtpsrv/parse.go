package smtpsrv

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// ParseSanitizeMessage takes a *mail.Message and returns a cleaned version of the Body
func ParseSanitizeMessage(m *mail.Message) ([]byte, error) {

	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		if err.Error() == "mime: no media type" {
			mediaType = "text/plain"
		} else {
			return nil, err
		}
	}

	switch i := true; i {

	case strings.HasPrefix(mediaType, "multipart/"):
		fmt.Println("case: multipart")
		body, err := parseMultipart(m.Body, params["boundary"])
		if err != nil {
			return nil, err
		}
		return body, nil

	case mediaType == "text/plain":
		fmt.Println("case: text/plain")
		body, err := io.ReadAll(m.Body)
		if err != nil {
			return nil, err
		}
		return body, nil

	case mediaType == "text/html":
		fmt.Println("case: text/html")
		body, err := io.ReadAll(m.Body)
		if err != nil {
			return nil, err
		}
		return parseHTML(body), nil

	default:
		fmt.Println("case: default")
		err := fmt.Errorf("unexpected content type: %s", mediaType)
		return nil, err
	}
}

// parseMultipart extracts the text/plain part from a multipart message
func parseMultipart(body io.Reader, boundary string) ([]byte, error) {
	mr := multipart.NewReader(body, boundary)
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		partType, _, err := mime.ParseMediaType(p.Header.Get("Content-Type"))
		if err != nil {
			if err.Error() == "mime: no media type" {
				partType = "text/plain"
			} else {
				return nil, err
			}
		}

		if partType == "text/plain" {
			slurp, err := io.ReadAll(p)
			if err != nil {
				return nil, err
			}
			return slurp, nil
		}
		if partType == "text/html" {
			slurp, err := io.ReadAll(p)
			if err != nil {
				return nil, err
			}
			return parseHTML(slurp), nil
		}
	}
	return nil, errors.New("multipart message missing a text/plain or text/html part")
}

// parseHTML returns the message body stripped of HTML tags
func parseHTML(body []byte) []byte {
	p := bluemonday.StripTagsPolicy()
	s := p.SanitizeBytes(body)
	return s
}
