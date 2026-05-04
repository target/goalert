package imapmanager

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/target/goalert/gadb"
)

// emailMessage represents a parsed email message.
type emailMessage struct {
	MessageID string
	From      string
	To        string
	Subject   string
	Body      string
	Date      string // Email date
	InReplyTo string // Message ID this email is replying to
}

// parseEnvelope extracts email fields from an IMAP message.
func parseEnvelope(msg *imap.Message) *emailMessage {
	if msg.Envelope == nil {
		return nil
	}

	em := &emailMessage{
		MessageID: msg.Envelope.MessageId,
		Subject:   msg.Envelope.Subject,
		Date:      msg.Envelope.Date.String(),
		InReplyTo: msg.Envelope.InReplyTo,
	}

	// Extract From address
	if len(msg.Envelope.From) > 0 {
		addr := msg.Envelope.From[0]
		if addr.MailboxName != "" && addr.HostName != "" {
			em.From = addr.MailboxName + "@" + addr.HostName
		}
	}

	// Extract To address(es)
	var toAddrs []string
	for _, addr := range msg.Envelope.To {
		if addr.MailboxName != "" && addr.HostName != "" {
			toAddrs = append(toAddrs, addr.MailboxName+"@"+addr.HostName)
		}
	}
	em.To = strings.Join(toAddrs, ", ")

	// Extract body (RFC822) - parse to get actual content without technical headers
	for _, bodyItem := range msg.Body {
		if bodyItem != nil {
			bodyBytes, err := io.ReadAll(bodyItem)
			if err == nil {
				// Parse the RFC822 message to extract just the body content
				em.Body = extractEmailBody(bodyBytes)
			}
		}
	}

	if em.MessageID == "" {
		return nil
	}

	return em
}

// extractEmailBody parses an RFC822 message and extracts just the text body content,
// stripping out all MIME headers (Delivered-To, Received, ARC-Seal, etc.).
func extractEmailBody(rfc822Data []byte) string {
	// Parse the RFC822 message
	msg, err := mail.ReadMessage(bytes.NewReader(rfc822Data))
	if err != nil {
		// If parsing fails, return empty string
		return ""
	}

	// Get the Content-Type header to check if it's multipart
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		// No Content-Type, try to read body directly
		bodyBytes, err := io.ReadAll(msg.Body)
		if err != nil {
			return ""
		}
		return string(bodyBytes)
	}

	// Parse the media type
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// If parsing fails, try to read body directly
		bodyBytes, err := io.ReadAll(msg.Body)
		if err != nil {
			return ""
		}
		return string(bodyBytes)
	}

	// Handle multipart messages
	if strings.HasPrefix(mediaType, "multipart/") {
		boundary, ok := params["boundary"]
		if !ok {
			return ""
		}

		mr := multipart.NewReader(msg.Body, boundary)
		var textBody, htmlBody string

		// Read all parts
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}

			partContentType := part.Header.Get("Content-Type")
			partMediaType, _, _ := mime.ParseMediaType(partContentType)

			partBytes, err := io.ReadAll(part)
			if err != nil {
				continue
			}

			// Prefer text/plain, fallback to text/html
			if partMediaType == "text/plain" {
				textBody = string(partBytes)
			} else if partMediaType == "text/html" && textBody == "" {
				htmlBody = string(partBytes)
			}
		}

		// Return text/plain if available, otherwise text/html
		if textBody != "" {
			return textBody
		}
		return htmlBody
	}

	// Not multipart, read body directly
	bodyBytes, err := io.ReadAll(msg.Body)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

// isReply checks if an email is a reply or forward based on subject and headers.
func isReply(email *emailMessage) bool {
	// Check subject line for common reply/forward prefixes
	subject := strings.TrimSpace(email.Subject)
	if len(subject) >= 3 {
		prefix := strings.ToLower(subject[:3])
		if prefix == "re:" || prefix == "fw:" {
			return true
		}
	}
	if len(subject) >= 4 {
		prefix := strings.ToLower(subject[:4])
		if prefix == "fwd:" {
			return true
		}
	}

	// Check In-Reply-To header
	if email.InReplyTo != "" {
		return true
	}

	return false
}

// matchesFilter checks if an email matches a filter rule.
func matchesFilter(email *emailMessage, rule gadb.IMAPFilterRulesForServiceRow) (bool, error) {
	// Check exclude_replies setting first
	if rule.ExcludeReplies && isReply(email) {
		return false, nil
	}

	// All criteria in a rule must match (AND logic)
	if rule.FromPattern.Valid {
		matched, err := matchPattern(email.From, rule.FromPattern.String, rule.MatchMode)
		if err != nil || !matched {
			return false, err
		}
	}

	if rule.SubjectPattern.Valid {
		matched, err := matchPattern(email.Subject, rule.SubjectPattern.String, rule.MatchMode)
		if err != nil || !matched {
			return false, err
		}
	}

	if rule.ToPattern.Valid {
		matched, err := matchPattern(email.To, rule.ToPattern.String, rule.MatchMode)
		if err != nil || !matched {
			return false, err
		}
	}

	return true, nil
}

// matchPattern performs pattern matching based on the match mode.
func matchPattern(value, pattern, mode string) (bool, error) {
	switch mode {
	case "exact":
		return strings.EqualFold(value, pattern), nil

	case "contains":
		return strings.Contains(strings.ToLower(value), strings.ToLower(pattern)), nil

	case "regex":
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, err
		}
		return re.MatchString(value), nil

	default:
		return false, nil
	}
}

