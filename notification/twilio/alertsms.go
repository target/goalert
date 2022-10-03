package twilio

import (
	"bytes"
	"context"
	"strings"
	"text/template"
	"unicode"

	"github.com/target/goalert/config"
	"github.com/target/goalert/notification"
	"github.com/target/goalert/util"
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

{{.Link}}{{end}}
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

// hasAnyPrefix returns true if any of the prefixes are present in the string.
func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if !strings.HasPrefix(s, p) {
			continue
		}

		return true
	}

	return false
}

// canContainURL returns true if the message can contain a URL.
func canContainURL(ctx context.Context, number string) bool {
	if config.FromContext(ctx).General.DisableSMSLinks {
		return false
	}

	return !hasAnyPrefix(number,
		// Non-exhaustive list of dialing codes that forbid URLs.
		"+86", // CN - https://www.twilio.com/guidelines/cn/sms
	)
}

// hasTwoWaySMSSupport returns true if a number supports 2-way SMS messaging (replies).
func hasTwoWaySMSSupport(ctx context.Context, number string) bool {
	if config.FromContext(ctx).Twilio.DisableTwoWaySMS {
		return false
	}

	return !hasAnyPrefix(number,
		// Non-exhaustive list of dialing codes that do not support 2-way SMS.
		"+91",  // IN - https://www.twilio.com/guidelines/in/sms
		"+86",  // CN - https://www.twilio.com/guidelines/cn/sms
		"+502", // GT - https://www.twilio.com/guidelines/gt/sms
		"+506", // CR - https://www.twilio.com/guidelines/cr/sms
		"+507", // PA - https://www.twilio.com/guidelines/pa/sms
		"+84",  // VN - https://www.twilio.com/guidelines/vn/sms
	)
}

func normalizeGSM(str string) (s string) {
	s = strings.Map(mapGSM, str)
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.TrimSpace(s)
	return s
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

	result, err := util.RenderSize(maxLen, data.Alert.Summary, func(summary string) (string, error) {
		buf.Reset()
		data.Alert.Summary = strings.TrimSpace(summary)
		err := alertTempl.Execute(&buf, data)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	})
	if err != nil {
		return "", err
	}

	return result, nil
}

// renderAlertStatusMessage will render a single-segment SMS for an Alert Status.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (if needed) until the output is <= maxLen characters.
func renderAlertStatusMessage(maxLen int, a notification.AlertStatus) (string, error) {
	var buf bytes.Buffer
	a.Summary = normalizeGSM(a.Summary)
	a.LogEntry = normalizeGSM(a.LogEntry)

	result, err := util.RenderSizeN(maxLen, []string{a.Summary, a.LogEntry}, func(inputs []string) (string, error) {
		buf.Reset()
		a.Summary = strings.TrimSpace(inputs[0])
		a.LogEntry = strings.TrimSpace(inputs[1])
		err := statusTempl.Execute(&buf, a)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	})
	if err != nil {
		return "", err
	}

	return result, nil
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

	result, err := util.RenderSize(maxLen, data.AlertBundle.ServiceName, func(name string) (string, error) {
		buf.Reset()
		data.AlertBundle.ServiceName = strings.TrimSpace(name)
		err := bundleTempl.Execute(&buf, data)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	})
	if err != nil {
		return "", err
	}

	return result, nil
}
