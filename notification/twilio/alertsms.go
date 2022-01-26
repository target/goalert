package twilio

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"text/template"
	"unicode"

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

var alertTempl = template.Must(template.New("alertSMS").Parse(`Alert #{{.AlertID}}: {{.Summary}}

{{- if .Link }}

{{.Link}}
{{end}}

{{- if .Code}}
Reply '{{.Code}}a' to ack, '{{.Code}}c' to close.{{end}}`))

var bundleTempl = template.Must(template.New("alertBundleSMS").Parse(`Svc '{{.ServiceName}}': {{.Count}} unacked alert{{if gt .Count 1}}s{{end}}

{{- if .Link }}

	{{.Link}}
{{end}}

{{- if .Code}}
	Reply '{{.Code}}aa' to ack all, '{{.Code}}cc' to close all.{{end}}`))

var statusTempl = template.Must(template.New("alertStatusSMS").Parse(`Alert #{{.AlertID}}{{- if .Summary }}: {{.Summary}}{{end}}

	{{.LogEntry}}`))

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

func normalizeGSM(str string) (s string) {
	s = strings.Map(mapGSM, str)
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.TrimSpace(s)
	return s
}

func trimString(str *string, buf *bytes.Buffer, maxLen int) bool {
	if buf.Len() <= maxLen {
		return false
	}

	newSumLen := len(*str) - (buf.Len() - maxLen)
	if newSumLen <= 0 {
		*str = ""
	}
	*str = strings.TrimSpace((*str)[:newSumLen])
	buf.Reset()

	return true
}

// renderAlertMessage will render a single-segment SMS for an Alert.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (if needed) until the output is <= maxLen characters.
func renderAlertMessage(maxLen int, a notification.Alert, link string, code int) (string, error) {
	var buf bytes.Buffer
	a.Summary = normalizeGSM(a.Summary)

	var data struct {
		notification.Alert
		Link string
		Code int
	}
	data.Alert = a
	data.Link = link
	data.Code = code

	err := alertTempl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	if trimString(&a.Summary, &buf, maxLen) {
		err = alertTempl.Execute(&buf, data)
		if err != nil {
			return "", err
		}
	}

	// should maybe revisit templates if this starts occurring
	if buf.Len() > maxLen {
		return "", errors.New("message too long")
	}

	return buf.String(), nil
}

// renderAlertStatusMessage will render a single-segment SMS for an Alert Status.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (if needed) until the output is <= maxLen characters.
func renderAlertStatusMessage(maxLen int, a notification.AlertStatus) (string, error) {
	var buf bytes.Buffer
	a.Summary = normalizeGSM(a.Summary)
	a.LogEntry = normalizeGSM(a.LogEntry)

	err := statusTempl.Execute(&buf, a)
	if err != nil {
		return "", err
	}

	if trimString(&a.Summary, &buf, maxLen) {
		err = statusTempl.Execute(&buf, a)
		if err != nil {
			return "", err
		}
	}

	if trimString(&a.LogEntry, &buf, maxLen) {
		err = statusTempl.Execute(&buf, a)
		if err != nil {
			return "", err
		}
	}

	// should maybe revisit templates if this starts occurring
	if buf.Len() > maxLen {
		return "", errors.New("message too long")
	}

	return buf.String(), nil
}

// renderAlertBundleMessage will render a single-segment SMS for an Alert Bundle.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (if needed) until the output is <= maxLen characters.
func renderAlertBundleMessage(maxLen int, a notification.AlertBundle, link string, code int) (string, error) {
	var buf bytes.Buffer
	a.ServiceName = normalizeGSM(a.ServiceName)

	var data struct {
		notification.AlertBundle
		Link string
		Code int
	}
	data.AlertBundle = a
	data.Link = link
	data.Code = code

	err := bundleTempl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	if trimString(&a.ServiceName, &buf, maxLen) {
		err = bundleTempl.Execute(&buf, data)
		if err != nil {
			return "", err
		}
	}

	// should maybe revisit templates if this starts occurring
	if buf.Len() > maxLen {
		return "", errors.New("message too long")
	}

	return buf.String(), nil
}
