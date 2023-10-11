package calsub

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// RFC can be found at https://tools.ietf.org/html/rfc5545
var iCalTemplate = template.Must(template.New("ical").Parse(strings.ReplaceAll(`BEGIN:VCALENDAR
PRODID:-//{{.ApplicationName}}//{{.Version}}//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
{{- $mins := .ReminderMinutes }}
{{- $genTime := .GeneratedAt }}
{{- $eventUIDs := .EventUIDs}}
{{- range $i, $s := .Shifts}}
BEGIN:VEVENT
UID:{{index $eventUIDs $i}}
SUMMARY:{{if $.FullSchedule}}{{index $.UserNames $s.UserID}} {{end}}On-Call ({{$.ApplicationName}}: {{$.ScheduleName}}){{if $s.Truncated}} Begins*
DESCRIPTION:The end time of this shift is unknown and will continue beyond what is displayed.
{{- end }}
DTSTAMP:{{$genTime.UTC.Format "20060102T150405Z"}}
DTSTART:{{.Start.UTC.Format "20060102T150405Z"}}
DTEND:{{.End.UTC.Format "20060102T150405Z"}}
{{- range $mins}}
BEGIN:VALARM
ACTION:DISPLAY
DESCRIPTION:REMINDER
TRIGGER:-PT{{.}}M
END:VALARM
{{- end}}
END:VEVENT
{{- end}}
END:VCALENDAR
`, "\n", "\r\n")))

// renderICal will generate an iCal file from the renderData.
func (r renderData) renderICal() ([]byte, error) {
	var icalRender struct {
		renderData
		EventUIDs []string
	}
	icalRender.renderData = r
	for _, s := range r.Shifts {
		t := s.End
		if s.Truncated {
			t = s.Start
		}
		sum := sha256.Sum256([]byte(s.UserID + r.ScheduleID.String() + t.Format(time.RFC3339)))
		icalRender.EventUIDs = append(icalRender.EventUIDs, hex.EncodeToString(sum[:]))
	}

	buf := bytes.NewBuffer(nil)
	err := iCalTemplate.Execute(buf, icalRender)
	if err != nil {
		return nil, errors.Wrap(err, "render ical template:")
	}

	return buf.Bytes(), nil
}
