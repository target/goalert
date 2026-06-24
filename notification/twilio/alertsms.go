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
	"github.com/target/goalert/util"
)

// 160 GSM characters (140 bytes) is the max for a single segment message.
// Multi-segment messages include a 6-byte header limiting to 153 GSM characters
// per segment.
//
// Non-GSM will use UCS-2 encoding, using 2-bytes per character. The max would
// then be 70 or 67 characters for single or multi-segmented messages, respectively.
const maxGSMLen = 160

var alertTempl = template.Must(template.New("alertSMS").Parse(`{{.AppName}}: Alert #{{.AlertID}}: {{.Summary}}
{{- if .Link }}

{{.Link}}{{end}}
{{- if .Code}}

Reply '{{.Code}}a' to ack, '{{.Code}}e' to escalate, '{{.Code}}c' to close.{{end}}`))

var bundleTempl = template.Must(template.New("alertBundleSMS").Parse(`{{.AppName}}: Svc '{{.ServiceName}}': {{.Count}} unacked alert{{if gt .Count 1}}s{{end}}

{{- if .Link }}

	{{.Link}}
{{end}}
{{- if .Code}}
	Reply '{{.Code}}aa' to ack all, '{{.Code}}cc' to close all.{{end}}`))

var statusTempl = template.Must(template.New("alertStatusSMS").Parse(`{{.AppName}}: Alert #{{.AlertID}}{{- if .Summary }}: {{.Summary}}{{end}}

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
	s = strings.ReplaceAll(s, "  ", " ")
	s = strings.TrimSpace(s)
	return s
}

func renderMinGSMSegments(inputs []string, render func(inputs []string) (string, error)) (string, error) {
	for i, s := range inputs {
		inputs[i] = normalizeGSM(s)
	}

	size := maxGSMLen
	for {
		result, err := util.RenderSizeN(size, inputs, func(inputs []string) (string, error) {
			cpy := make([]string, len(inputs))
			for i, s := range inputs {
				cpy[i] = strings.TrimSpace(s)
			}

			return render(cpy)
		})
		if errors.Is(err, util.ErrNoSolution) {
			size += maxGSMLen
			continue
		}
		if err != nil {
			return "", err
		}

		return result, nil
	}
}

// renderAlertMessage will render a SMS message for an Alert.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (as needed) to use the minimum number of message segments.
func renderAlertMessage(appName string, a notification.Alert, link string, code int) (string, error) {
	var buf bytes.Buffer
	var data struct {
		AppName string
		notification.Alert
		Link string
		Code int
	}
	data.AppName = appName
	data.Alert = a
	data.Link = link
	data.Code = code

	result, err := renderMinGSMSegments([]string{a.Summary}, func(inputs []string) (string, error) {
		buf.Reset()
		data.Summary = inputs[0]
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

// renderAlertStatusMessage will render a SMS message for an Alert Status.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (as needed) to use the minimum number of message segments.
func renderAlertStatusMessage(appName string, a notification.AlertStatus) (string, error) {
	var buf bytes.Buffer
	var data struct {
		AppName string
		notification.AlertStatus
	}
	data.AppName = appName
	data.AlertStatus = a
	result, err := renderMinGSMSegments([]string{a.Summary, a.LogEntry}, func(inputs []string) (string, error) {
		buf.Reset()
		data.Summary = inputs[0]
		data.LogEntry = inputs[1]
		err := statusTempl.Execute(&buf, data)
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

// renderAlertBundleMessage will render an SMS message for an Alert Bundle.
//
// Non-GSM characters will be replaced with '?' and fields will be
// truncated (as needed) to use the minimum number of message segments.
func renderAlertBundleMessage(appName string, a notification.AlertBundle, link string, code int) (string, error) {
	var buf bytes.Buffer

	var data struct {
		AppName string
		notification.AlertBundle
		Link string
		Code int
	}
	data.AppName = appName
	data.AlertBundle = a
	data.Link = link
	data.Code = code

	result, err := renderMinGSMSegments([]string{data.ServiceName}, func(inputs []string) (string, error) {
		buf.Reset()
		data.ServiceName = inputs[0]
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
