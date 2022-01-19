package twilio

import (
	"bytes"
	"context"
	"strings"
	"text/template"
	"unicode"

	"github.com/pkg/errors"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
)

// 160 GSM characters (140 bytes) is the max for a single segment message.
// Multi-segment messages include a 6-byte header limiting to 153 GSM characters
// per segment.
//
// Non-GSM will use UCS-2 encoding, using 2-bytes per character. The max would
// then be 70 or 67 characters for single or multi-segmented messages, respectively.
const maxGSMLen = 160

type alertSMS struct {
	ID      int
	Count   int
	Body    string
	Summary string
	Link    string
	Code    int
	Type    notification.MessageType
}

// alertTempl uses the ID, Summary, Link, and Code to render a message
var alertTempl = template.Must(template.New("alertSMS").Parse(`Alert #{{.ID}}: {{.Summary}}

{{- if .Link }}
{{.Link}}
{{end}}

{{- if .Code}}
Reply '{{.Code}}a' to ack, '{{.Code}}c' to close.{{end}}`))

// bundleTempl uses the Count, Body, Link, and Code to render a message
var bundleTempl = template.Must(template.New("alertBundleSMS").Parse(`Svc '{{.Body}}': {{.Count}} unacked alert{{if gt .Count 1}}s{{end}}

{{- if .Link }}
{{.Link}}
{{end}}

{{- if .Code}}
Reply '{{.Code}}aa' to ack all, '{{.Code}}cc' to close all.{{end}}`))

// statusTempl uses the ID, Summary, and Body to render a message
var statusTempl = template.Must(template.New("alertStatusSMS").Parse(`Alert #{{.ID}}{{-if .Summary }}: {{.Summary}}{{end}}

{{.Body}}`))

const gsmAlphabet = "@∆ 0¡P¿p£!1AQaq$Φ\"2BRbr¥Γ#3CScsèΛ¤4DTdtéΩ%5EUeuùΠ&6FVfvìΨ'7GWgwòΣ(8HXhxÇΘ)9IYiy\n Ξ *:JZjzØ+;KÄkäøÆ,<LÖlö\ræ-=MÑmñÅß.>NÜnüåÉ/?O§oà"

var gsmChr = make(map[rune]bool, len(gsmAlphabet))

func init() {
	for _, r := range gsmAlphabet {
		gsmChr[r] = true
	}
}

func mapGSM(r rune) rune {
	if unicode.IsSpace(r) {
		return ' '
	}

	if !unicode.IsPrint(r) {
		return -1
	}

	if gsmChr[r] {
		return r
	}

	// Map similar characters to keep as much meaning as possible.
	switch r {
	case '_', '|', '~':
		return '-'
	case '[', '{':
		return '('
	case ']', '}':
		return ')'
	case '»':
		return '>'
	case '`', '’', '‘':
		return '\''
	}

	switch {
	case unicode.Is(unicode.Dash, r):
		return '-'
	case unicode.Is(unicode.Quotation_Mark, r):
		return '"'
	}

	// If no substitute, replace with '?'
	return '?'
}

// hasTwoWaySMSSupport returns true if a number supports 2-way SMS messaging (replies).
func hasTwoWaySMSSupport(ctx context.Context, number string) bool {
	if config.FromContext(ctx).Twilio.DisableTwoWaySMS {
		return false
	}

	// India numbers do not support SMS replies.
	return !strings.HasPrefix(number, "+91")
}

// Render will render a single-segment SMS.
//
// Non-GSM characters will be replaced with '?' and Body will be
// truncated (if needed) until the output is <= maxLen characters.
func (a alertSMS) Render(maxLen int) (string, error) {
	a.Body = strings.Map(mapGSM, a.Body)
	a.Body = strings.Replace(a.Body, "  ", " ", -1)
	a.Body = strings.TrimSpace(a.Body)

	a.Summary = strings.Map(mapGSM, a.Summary)
	a.Summary = strings.Replace(a.Summary, "  ", " ", -1)
	a.Summary = strings.TrimSpace(a.Summary)

	var buf bytes.Buffer

	var tmpl template.Template
	switch a.Type {
	case notification.MessageTypeAlertStatus:
		tmpl = *statusTempl
	case notification.MessageTypeAlertBundle:
		tmpl = *bundleTempl
	case notification.MessageTypeAlert:
		tmpl = *alertTempl
	default:
		return "", errors.Errorf("unsupported message type: %v", a.Type)
	}

	err := tmpl.Execute(&buf, a)
	if err != nil {
		return "", err
	}

	if buf.Len() > maxLen {
		// message too long, trim summary (until empty if needed)
		newSumLen := len(a.Summary) - (buf.Len() - maxLen)
		if newSumLen <= 0 {
			a.Summary = ""
		}
		a.Summary = strings.TrimSpace(a.Summary[:newSumLen])
		buf.Reset()
		err = tmpl.Execute(&buf, a)
		if err != nil {
			return "", err
		}
	}

	if buf.Len() > maxLen {
		// trim body of message if message still too long
		newBodyLen := len(a.Body) - (buf.Len() - maxLen)
		if newBodyLen <= 0 {
			return "", errors.New("message too long to include body")
		}
		a.Body = strings.TrimSpace(a.Body[:newBodyLen])
		buf.Reset()
		err = tmpl.Execute(&buf, a)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
