package twilio

import (
	"bytes"
	"strings"
	"text/template"
	"unicode"

	"github.com/pkg/errors"
)

// 160 GSM characters (140 bytes) is the max for a single segment message.
// Multi-segment messages include a 6-byte header limiting to 153 GSM characters
// per segment.
//
// Non-GSM will use UCS-2 encoding, using 2-bytes per character. The max would
// then be 70 or 67 characters for single or multi-segmented messages, respectively.
const maxGSMLen = 160

type alertSMS struct {
	ID   int
	Body string
	Link string
	Code int
}

var smsTmpl = template.Must(template.New("alertSMS").Parse(
	`Alert #{{.ID}}: {{.Body}}
{{- if .Link }}

{{.Link}}
{{- end}}
{{- if .Code}}

Reply '{{.Code}}a' to ack, '{{.Code}}c' to close.
{{- end}}`,
))

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
func hasTwoWaySMSSupport(number string) bool {
	// India numbers do not support SMS replies.
	return !strings.HasPrefix(number, "+91")
}

// Render will render a single-segment SMS.
//
// Non-GSM characters will be replaced with '?' and Body will be
// truncated (if needed) until the output is <= 160 characters.
func (a alertSMS) Render() (string, error) {
	a.Body = strings.Map(mapGSM, a.Body)
	a.Body = strings.Replace(a.Body, "  ", " ", -1)
	a.Body = strings.TrimSpace(a.Body)

	var buf bytes.Buffer
	err := smsTmpl.Execute(&buf, a)
	if err != nil {
		return "", err
	}

	if buf.Len() > maxGSMLen {
		newBodyLen := len(a.Body) - (buf.Len() - maxGSMLen)
		if newBodyLen <= 0 {
			return "", errors.New("message too long to include body")
		}
		a.Body = strings.TrimSpace(a.Body[:newBodyLen])
		buf.Reset()
		err = smsTmpl.Execute(&buf, a)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
