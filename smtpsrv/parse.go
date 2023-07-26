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

	switch {
	case strings.HasPrefix(mediaType, "multipart/"):
		return parseMultipart(m.Body, params["boundary"])
	case mediaType == "text/plain":
		return io.ReadAll(m.Body)
	case mediaType == "text/html":
		body, err := io.ReadAll(m.Body)
		if err != nil {
			return nil, err
		}
		return parseHTML(body), nil
	}

	return nil, fmt.Errorf("unexpected content type: %s", mediaType)
}

// parseMultipart extracts the text from a multipart message.
func parseMultipart(body io.Reader, boundary string) ([]byte, error) {
	mr := multipart.NewReader(body, boundary)
	var htmlData []byte
	var hasHTML bool
	for {
		p, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		ct := p.Header.Get("Content-Type")
		if ct == "" {
			ct = "text/plain"
		}

		partType, _, err := mime.ParseMediaType(ct)
		if err != nil {
			return nil, err
		}
		switch partType {
		case "text/plain":
			return io.ReadAll(p)
		case "text/html":
			hasHTML = true
			htmlData, err = io.ReadAll(p)
			if err != nil {
				return nil, err
			}
		}
	}
	if !hasHTML {
		return nil, errors.New("multipart message missing a text/plain or text/html part")
	}

	return parseHTML(htmlData), nil
}

// parseHTML returns the message body stripped of HTML tags
func parseHTML(body []byte) []byte {
	p := bluemonday.StripTagsPolicy()
	s := p.SanitizeBytes(body)
	return s
}
